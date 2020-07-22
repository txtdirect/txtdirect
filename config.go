package txtdirect

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"gopkg.in/natefinch/lumberjack.v2"
)

var allOptions = []string{"host", "path", "gometa", "www"}

// Config contains the middleware's configuration
type Config struct {
	Enable    []string `json:"enable"`
	Redirect  string   `json:"redirect,omitempty"`
	Resolver  string   `json:"resolver,omitempty"`
	LogOutput string   `json:"logfile,omitempty"`
	Qr        Qr
}

func ParseCaddy(d *caddyfile.Dispenser) (*Config, error) {
	var enable []string
	var redirect string
	var resolver string
	var logfile string

	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "disable":
				if enable != nil {
					return nil, d.ArgErr()
				}
				toDisable := d.RemainingArgs()
				if len(toDisable) == 0 {
					return nil, d.ArgErr()
				}
				enable = removeArrayFromArray(allOptions, toDisable)

			case "enable":
				if enable != nil {
					return nil, d.ArgErr()
				}
				enable = d.RemainingArgs()
				if len(enable) == 0 {
					return nil, d.ArgErr()
				}

			case "redirect":
				toRedirect := d.RemainingArgs()
				if len(toRedirect) != 1 {
					return nil, d.ArgErr()
				}
				redirect = toRedirect[0]

			case "resolver":
				resolverAddr := d.RemainingArgs()
				if len(resolverAddr) != 1 {
					return nil, d.ArgErr()
				}
				resolver = resolverAddr[0]

			case "logfile":
				logfile = "stdout"
				// Set stdout as the default value
				if d.NextArg() {
					logfile = d.Val()
				}
			}
		}

	}

	// If nothing is specified, enable everything
	if enable == nil {
		enable = allOptions
	}

	conf := Config{
		Enable:    enable,
		Redirect:  redirect,
		Resolver:  resolver,
		LogOutput: logfile,
	}

	parseLogfile(logfile)

	return &conf, nil
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
