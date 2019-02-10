/*
Copyright 2017 - The TXTdirect Authors
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

package caddy

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/txtdirect/txtdirect"
)

func init() {
	caddy.RegisterPlugin("txtdirect", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

var allOptions = []string{"host", "path", "gometa", "www"}

func parse(c *caddy.Controller) (txtdirect.Config, error) {
	var enable []string
	var redirect string
	var resolver string
	var gomods txtdirect.Gomods
	var prometheus txtdirect.Prometheus

	c.Next() // skip directive name
	for c.NextBlock() {
		option := c.Val()
		switch option {
		case "disable":
			if enable != nil {
				return txtdirect.Config{}, c.ArgErr()
			}
			toDisable := c.RemainingArgs()
			if len(toDisable) == 0 {
				return txtdirect.Config{}, c.ArgErr()
			}
			enable = removeArrayFromArray(allOptions, toDisable)

		case "enable":
			if enable != nil {
				return txtdirect.Config{}, c.ArgErr()
			}
			enable = c.RemainingArgs()
			if len(enable) == 0 {
				return txtdirect.Config{}, c.ArgErr()
			}

		case "redirect":
			toRedirect := c.RemainingArgs()
			if len(toRedirect) != 1 {
				return txtdirect.Config{}, c.ArgErr()
			}
			redirect = toRedirect[0]

		case "resolver":
			resolverAddr := c.RemainingArgs()
			if len(resolverAddr) != 1 {
				return txtdirect.Config{}, c.ArgErr()
			}
			resolver = resolverAddr[0]

		case "logfile":
			logfile := c.RemainingArgs()
			if len(logfile) != 1 {

			}
			parseLogfile(logfile[0])
		case "gomods":
			gomods.Enable = true
			c.NextArg()
			if c.Val() != "{" {
				continue
			}
			for c.Next() {
				if c.Val() == "}" {
					break
				}
				err := gomods.ParseGomods(c)
				if err != nil {
					return txtdirect.Config{}, err
				}
			}

		case "prometheus":
			prometheus.Enable = true
			c.NextArg()
			if c.Val() != "{" {
				continue
			}
			for c.Next() {
				if c.Val() == "}" {
					break
				}
				err := prometheus.ParsePrometheus(c, c.Val(), c.RemainingArgs()[0])
				if err != nil {
					return txtdirect.Config{}, err
				}
			}

		default:
			return txtdirect.Config{}, c.ArgErr() // unhandled option
		}
	}

	// If nothing is specified, enable everything
	if enable == nil {
		enable = allOptions
	}

	if gomods.Enable == true {
		gomods.SetDefaults()
	}
	if prometheus.Enable == true {
		prometheus.SetDefaults()
	}

	config := txtdirect.Config{
		Enable:     enable,
		Redirect:   redirect,
		Resolver:   resolver,
		Gomods:     gomods,
		Prometheus: prometheus,
	}

	return config, nil
}

func setup(c *caddy.Controller) error {
	config, err := parse(c)
	if err != nil {
		return err
	}

	// Add handler to Caddy
	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return Redirect{
			Next:   next,
			Config: config,
		}
	}
	cfg.AddMiddleware(mid)
	if config.Prometheus.Enable {
		config.Prometheus.SetDefaults()
		config.Prometheus.Setup(c)
	}

	return nil
}

func removeArrayFromArray(array, toBeRemoved []string) []string {
	tmp := make([]string, len(array))
	copy(tmp, array)
	for _, toRemove := range toBeRemoved {
		for i, option := range tmp {
			if option == toRemove {
				tmp[i] = tmp[len(tmp)-1]
				tmp = tmp[:len(tmp)-1]
				break
			}
		}
	}
	return tmp
}

// Redirect is middleware to redirect requests based on TXT records
type Redirect struct {
	Next   httpserver.Handler
	Config txtdirect.Config
}

func (rd Redirect) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if err := txtdirect.Redirect(w, r, rd.Config); err != nil {
		if err.Error() == "option disabled" {
			return rd.Next.ServeHTTP(w, r)
		}
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

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
