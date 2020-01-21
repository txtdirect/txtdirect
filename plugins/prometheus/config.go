package prometheus

import (
	"net/http"
	"strconv"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable  bool
	Address string
	Path    string

	Next    httpserver.Handler
	Handler http.Handler
}

// ParsePrometheus parses the txtdirect config for Prometheus
func (p *Prometheus) ParsePrometheus(c *caddy.Controller, key, value string) error {
	switch key {
	case "enable":
		value, err := strconv.ParseBool(value)
		if err != nil {
			return c.ArgErr()
		}
		p.Enable = value
	case "address":
		// TODO: validate the given address
		p.Address = value
	case "path":
		p.Path = value
	default:
		return c.ArgErr() // unhandled option for prometheus
	}
	return nil
}
