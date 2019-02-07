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
	"github.com/gomods/athens/pkg/paths"
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
	Fs       afero.Fs
}

type Cache struct {
	Enable bool
	Type   string
	Path   string
}

type Module struct {
	Name    string
	Version string
	FileExt string
}

type ModuleHandler interface {
	fetch(r *http.Request, c Config) (*storage.Version, error)
	storage(c Config) (storage.Backend, error)
	dp(fetcher module.Fetcher, s storage.Backend, fs afero.Fs) download.Protocol
}

var gomodsRegex = regexp.MustCompile("(list|info|mod|zip)")
var modVersionRegex = regexp.MustCompile("(.*)\\.(info|mod|zip)")
var DefaultGoBinaryPath = os.Getenv("GOROOT") + "/bin/go"

const (
	DefaultGomodsCacheType = "tmp"
	DefaultGomodsWorkers   = 1
)

// SetDefaults sets the default values for gomods config
// if the fields are empty
func (gomods *Gomods) SetDefaults() {
	gomods.Fs = afero.NewOsFs()
	if gomods.GoBinary == "" {
		gomods.GoBinary = DefaultGoBinaryPath
	}
	if gomods.Cache.Enable {
		if gomods.Cache.Type == "" {
			gomods.Cache.Type = DefaultGomodsCacheType
		}
		if gomods.Cache.Path == "" {
			gomods.Cache.Path = afero.GetTempDir(gomods.Fs, "")
		}
	}
	if gomods.Workers == 0 {
		gomods.Workers = DefaultGomodsWorkers
	}
}

func gomods(w http.ResponseWriter, r *http.Request, path string, c Config) error {
	m := Module{}
	if err := m.ParseImportPath(path); err != nil {
		return fmt.Errorf("module url is empty")
	}

	dp, err := m.fetch(r, c)
	if err != nil {
		return err
	}

	switch m.FileExt {
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
	fetcher, err := module.NewGoGetFetcher(c.Gomods.GoBinary, c.Gomods.Fs)
	if err != nil {
		return nil, err
	}
	s, err := m.storage(c)
	if err != nil {
		return nil, err
	}
	dp := m.dp(fetcher, s, c)
	return dp, nil
}

func (m Module) storage(c Config) (storage.Backend, error) {
	switch c.Gomods.Cache.Type {
	case "local":
		// Check if cache storage path exists, if not create it
		if _, err := os.Stat(c.Gomods.Cache.Path); os.IsNotExist(err) {
			if err = os.MkdirAll(c.Gomods.Cache.Path, os.ModePerm); err != nil {
				return nil, fmt.Errorf("couldn't create the cache storage directory on %s: %s", c.Gomods.Cache.Path, err.Error())
			}
		}
		s, err := fs.NewStorage(c.Gomods.Cache.Path, afero.NewOsFs())
		if err != nil {
			return nil, fmt.Errorf("could not create new storage from os fs (%s)", err)
		}
		return s, nil
	case "tmp":
		s, err := fs.NewStorage(c.Gomods.Cache.Path, afero.NewOsFs())
		if err != nil {
			return nil, fmt.Errorf("could not create new storage from os fs (%s)", err)
		}
		return s, nil
	}
	return nil, fmt.Errorf("Invalid storage config for gomods")
}

func (m Module) dp(fetcher module.Fetcher, s storage.Backend, c Config) download.Protocol {
	lister := download.NewVCSLister(c.Gomods.GoBinary, c.Gomods.Fs)
	st := stash.New(fetcher, s, stash.WithPool(c.Gomods.Workers), stash.WithSingleflight)
	dpOpts := &download.Opts{
		Storage: s,
		Stasher: st,
		Lister:  lister,
	}
	dp := download.New(dpOpts, addons.WithPool(c.Gomods.Workers))
	return dp
}

// ParseImportPath parses the request path and exports the
// module's import path, module's version and file extension
func (m *Module) ParseImportPath(path string) error {
	if strings.Contains(path, "@latest") {
		pathLatest := strings.Split(path, "/@")
		m.Name, m.Version, m.FileExt = pathLatest[0][1:], "", pathLatest[1]
		if err := m.DecodeImportPath(); err != nil {
			return err
		}
		return nil
	}

	// First item in array is modules import path and the secondd item is version+extension
	pathSlice := strings.Split(path, "/@v/")
	if pathSlice[1] == "list" {
		m.Name, m.Version, m.FileExt = pathSlice[0][1:], "", "list"
		if err := m.DecodeImportPath(); err != nil {
			return err
		}
		return nil
	}

	versionExt := modVersionRegex.FindAllStringSubmatch(pathSlice[1], -1)[0]
	m.Name, m.Version, m.FileExt = pathSlice[0][1:], versionExt[1], versionExt[2]
	if err := m.DecodeImportPath(); err != nil {
		return err
	}

	return nil
}

// DecodeImportPath decodes the module's import path. For more information check
// https://github.com/golang/go/blob/master/src/cmd/go/internal/module/module.go#L375-L433
func (m *Module) DecodeImportPath() error {
	decoded, err := paths.DecodePath(m.Name)
	if err != nil {
		return err
	}
	m.Name = decoded
	return nil
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
