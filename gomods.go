package txtdirect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/download/addons"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/mholt/caddy"
	"github.com/spf13/afero"
)

type Gomods struct {
	Enable   bool
	GoBinary string
	Workers  int
	Cache    Cache
}
type Cache struct {
	Enable bool
	Type   string
	Path   string
}

type Module struct {
	Name    string
	Version string
}

type ModuleHandler interface {
	fetch(r *http.Request, c Config) (*storage.Version, error)
	storage(c Config) (storage.Backend, error)
	dp(fetcher module.Fetcher, s storage.Backend, fs afero.Fs) download.Protocol
}

var gomodsRegex = regexp.MustCompile("(list|info|mod|zip)")
var modVersionRegex = regexp.MustCompile("@v\\/(v\\d+\\.\\d+\\.\\d+\\-\\d+\\-[\\w\\d]+|v\\d+\\.\\d+\\.\\d+|v\\d+\\.\\d+|latest|master)")
var DefaultGoBinaryPath = os.Getenv("GOROOT") + "/bin/go"

const (
	DefaultGomodsCachePath = "/tmp/txtdirect/gomods"
	DefaultGomodsCacheType = "tmp"
	DefaultGomodsWorkers   = 1
)

// SetDefaults sets the default values for gomods config
// if the fields are empty
func (gomods *Gomods) SetDefaults() {
	if gomods.GoBinary == "" {
		gomods.GoBinary = DefaultGoBinaryPath
	}
	if gomods.Cache.Enable {
		if gomods.Cache.Type == "" {
			gomods.Cache.Type = DefaultGomodsCacheType
		}
		if gomods.Cache.Path == "" {
			gomods.Cache.Path = DefaultGomodsCachePath
		}
	}
	if gomods.Workers == 0 {
		gomods.Workers = DefaultGomodsWorkers
	}
}

func gomods(w http.ResponseWriter, r *http.Request, path string, c Config) error {
	moduleName, version, ext := moduleNameAndVersion(path)
	if moduleName == "" {
		return fmt.Errorf("module url is empty")
	}
	m := Module{
		Name:    moduleName,
		Version: version,
	}
	dp, err := m.fetch(r, c)
	if err != nil {
		return err
	}
	switch ext {
	case "list":
		list, err := dp.List(r.Context(), m.Name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(strings.Join(list, "\n")))
		if err != nil {
			return err
		}
		return nil
	case "info":
		info, err := dp.Info(r.Context(), m.Name, m.Version)
		if err != nil {
			return err
		}
		_, err = w.Write(info)
		if err != nil {
			return err
		}
		return nil
	case "mod":
		mod, err := dp.GoMod(r.Context(), m.Name, m.Version)
		if err != nil {
			return err
		}
		_, err = w.Write(mod)
		if err != nil {
			return err
		}
		return nil
	case "zip":
		zip, err := dp.Zip(r.Context(), m.Name, m.Version)
		if err != nil {
			return err
		}
		defer zip.Close()
		w.Write([]byte{})
		_, err = io.Copy(w, zip)
		if err != nil {
			return err
		}
		return nil
	case "latest":
		info, err := dp.Latest(r.Context(), m.Name)
		if err != nil {
			return err
		}
		json, err := json.Marshal(info)
		if err != nil {
			return err
		}
		_, err = w.Write(json)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("the requested file's extension is not supported")
	}
}

func (m Module) fetch(r *http.Request, c Config) (download.Protocol, error) {
	fs := afero.NewOsFs()
	fetcher, err := module.NewGoGetFetcher(c.Gomods.GoBinary, fs)
	if err != nil {
		return nil, err
	}
	s, err := m.storage(c)
	if err != nil {
		return nil, err
	}
	dp := m.dp(fetcher, s, fs, c)
	return dp, nil
}

func (m Module) storage(c Config) (storage.Backend, error) {
	switch c.Gomods.Cache.Type {
	case "local":
		s, err := fs.NewStorage(c.Gomods.Cache.Path, afero.NewOsFs())
		if err != nil {
			return nil, fmt.Errorf("could not create new storage from os fs (%s)", err)
		}
		return s, nil
	}
	return nil, fmt.Errorf("Invalid storage config for gomods")
}

func (m Module) dp(fetcher module.Fetcher, s storage.Backend, fs afero.Fs, c Config) download.Protocol {
	lister := download.NewVCSLister(c.Gomods.GoBinary, fs)
	st := stash.New(fetcher, s, stash.WithPool(c.Gomods.Workers), stash.WithSingleflight)
	dpOpts := &download.Opts{
		Storage: s,
		Stasher: st,
		Lister:  lister,
	}
	dp := download.New(dpOpts, addons.WithPool(c.Gomods.Workers))
	return dp
}

func moduleNameAndVersion(path string) (string, string, string) {
	pathSlice := strings.Split(path, "/")[1:] // [1:] ignores the empty slice item
	var ext, version string
	for k, v := range pathSlice {
		if v == "@v" {
			ext = gomodsRegex.FindAllStringSubmatch(pathSlice[k+1], -1)[0][0]
			break
		}
	}
	moduleName := strings.Join(pathSlice[0:3], "/")
	if strings.Contains(path, "@latest") {
		return strings.Join(pathSlice[0:3], "/"), "", "latest"
	}
	if !strings.Contains(path, "list") {
		version = modVersionRegex.FindAllStringSubmatch(path, -1)[0][1]
	}
	return moduleName, version, ext
}

// ParseGomods parses the txtdirect config for gomods
func (gomods *Gomods) ParseGomods(c *caddy.Controller) error {
	switch c.Val() {
	case "gobinary":
		gomods.GoBinary = c.RemainingArgs()[0]

	case "workers":
		value, err := strconv.Atoi(c.RemainingArgs()[0])
		if err != nil {
			return c.ArgErr()
		}
		gomods.Workers = value

	case "cache":
		gomods.Cache.Enable = true
		c.NextArg()
		if c.Val() != "{" {
			break
		}
		for c.Next() {
			if c.Val() == "}" {
				break
			}
			err := gomods.Cache.ParseCache(c)
			if err != nil {
				return err
			}
		}
	default:
		return c.ArgErr() // unhandled option for gomods
	}
	return nil
}

// ParseCache parses the txtdirect config for gomods cache
func (cache *Cache) ParseCache(c *caddy.Controller) error {
	switch c.Val() {
	case "type":
		cache.Type = c.RemainingArgs()[0]
	case "path":
		cache.Path = c.RemainingArgs()[0]
	default:
		return c.ArgErr() // unhandled option for gomods cache
	}
	return nil
}
