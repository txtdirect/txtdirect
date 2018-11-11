package txtdirect

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/txtdirect/txtdirect/pkg/cache"

	"github.com/mholt/caddy/caddyhttp/proxy"
)

type ModProxy struct {
	Enable bool
	Path   string
	Cache  struct {
		Enable bool
		Type   string
		Path   string
	}
}

type Module struct {
	Path      string
	Version   string
	LocalPath string
}

type ModuleHandler interface {
	proxy() error
	cache() error
	zip() error
}

func gomods(w http.ResponseWriter, r *http.Request, path string, c Config) error {
	pathSlice := strings.Split(path, "/")[1:] // [1:] ignores the empty slice item
	var moduleName string
	var fileName string
	for k, v := range pathSlice {
		if v == "@v" {
			fileName = pathSlice[k+1]
			break
		}
		moduleName = strings.Join([]string{moduleName, v}, "/")
	}
	localPath := fmt.Sprintf("%s/%s", c.ModProxy.Cache.Path, moduleName[1:])
	m := Module{
		Path:      moduleName[1:], // [1:] ignores "/" at the beginning of url
		LocalPath: localPath,
		Version:   strings.Split(fileName, ".")[0], // Gets version number from last part of the path
	}
	u, err := url.Parse(fmt.Sprintf("https://%s/@v/%s", m.Path, fileName))
	if err != nil {
		return fmt.Errorf("unable to parse the url: %s", err.Error())
	}
	if c.ModProxy.Cache.Enable {
		err = m.cache(u, c)
		if err != nil {
			return fmt.Errorf("unable to cache the file on %s storage: %s", c.ModProxy.Cache.Type, err.Error())
		}
	}
	err = m.proxy(w, r, u)
	if err != nil {
		return fmt.Errorf("unable to proxy the request: %s", err.Error())
	}

	return nil
}

func (m Module) proxy(w http.ResponseWriter, r *http.Request, u *url.URL) error {
	r.URL.Path = "" // FIXME: Reconsider this part
	reverseProxy := proxy.NewSingleHostReverseProxy(u, "", proxyKeepalive, proxyTimeout, fallbackDelay)
	if err := reverseProxy.ServeHTTP(w, r, nil); err != nil {
		return err
	}
	return nil
}

func (m Module) cache(u *url.URL, c Config) error {
	switch c.ModProxy.Cache.Type {
	case "local":
		if err := cache.Local(m.Path, m.LocalPath, m.Version); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unable to identify the %s storage", c.ModProxy.Cache.Type)
	}
	return nil
}
