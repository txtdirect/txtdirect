/*
Copyright 2017 - The TXTDirect Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"go.txtdirect.org/txtdirect"
)

func main() {
	caddycmd.Main()
}

func init() {
	caddy.RegisterModule(TXTDirect{})
	httpcaddyfile.RegisterHandlerDirective("txtdirect", parseCaddyfile)
}

// TXTDirect implements an HTTP handler that calls
// Redirect method on requests and writes back the response
type TXTDirect struct {
	*txtdirect.Config
}

// CaddyModule returns the Caddy module information.
func (TXTDirect) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.txtdirect",
		New: func() caddy.Module { return new(TXTDirect) },
	}
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (t TXTDirect) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	r.URL.Host = strings.ToLower(r.URL.Host)
	if err := txtdirect.Redirect(w, r, *t.Config); err != nil {
		if err.Error() == "option disabled" {
			return next.ServeHTTP(w, r)
		}
		return err
	}

	// Count total redirects if prometheus is enabled and set cache header
	if w.Header().Get("Status-Code") == "301" || w.Header().Get("Status-Code") == "302" {
		// Set Cache-Control header on permanent redirects
		if w.Header().Get("Status-Code") == "301" {
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", txtdirect.Status301CacheAge))
		}
	}

	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (t *TXTDirect) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	var err error
	if t.Config, err = txtdirect.ParseCaddy(d); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't parse the config: %s", err.Error())
	}

	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var t TXTDirect
	err := t.UnmarshalCaddyfile(h.Dispenser)
	return t, err
}

// Interface guards
var (
	_ caddyhttp.MiddlewareHandler = (*TXTDirect)(nil)
	_ caddyfile.Unmarshaler       = (*TXTDirect)(nil)
)

func parseLogfile(logfile string) {
	switch logfile {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}
}
