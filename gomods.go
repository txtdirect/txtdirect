package txtdirect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/download/addons"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/spf13/afero"
)

type ModProxy struct {
	Enable   bool
	Path     string
	GoBinary string
	Cache    struct {
		Type string
		Path string
	}
}

type Module struct {
	Name    string
	Version string
}

type ModuleHandler interface {
	fetch() (*storage.Version, error)
	storage(c Config) (storage.Backend, error)
	dp(fetcher module.Fetcher, s storage.Backend, fs afero.Fs) download.Protocol
}

var gomodsRegex = regexp.MustCompile("(list|info|mod|zip)")
var modVersionRegex = regexp.MustCompile("@v\\/(v\\d+\\.\\d+\\.\\d+?|v\\d+\\.\\d+|latest|master)")

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

	return nil
}

func (m Module) fetch(r *http.Request, c Config) (download.Protocol, error) {
	fs := afero.NewOsFs()
	fetcher, err := module.NewGoGetFetcher("/usr/local/go/bin/go", fs)
	if err != nil {
		return nil, err
	}
	s, err := m.storage(c)
	if err != nil {
		return nil, err
	}
	dp := m.dp(fetcher, s, fs)
	return dp, nil
}

func (m Module) storage(c Config) (storage.Backend, error) {
	switch c.ModProxy.Cache.Type {
	case "local":
		s, err := fs.NewStorage(c.ModProxy.Cache.Path, afero.NewOsFs())
		if err != nil {
			return nil, fmt.Errorf("could not create new storage from os fs (%s)", err)
		}
		return s, nil
	}
	return nil, fmt.Errorf("Invalid storage config for gomods")
}

func (m Module) dp(fetcher module.Fetcher, s storage.Backend, fs afero.Fs) download.Protocol {
	lister := download.NewVCSLister("/usr/local/go/bin/go", fs)
	st := stash.New(fetcher, s, stash.WithPool(2), stash.WithSingleflight)
	dpOpts := &download.Opts{
		Storage: s,
		Stasher: st,
		Lister:  lister,
	}
	dp := download.New(dpOpts, addons.WithPool(2))
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
