package config

import (
	"net/http"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"go.txtdirect.org/txtdirect/plugins/qr"
)

// Config contains the middleware's configuration
type Config struct {
	Enable    []string
	Redirect  string
	Resolver  string
	LogOutput string
	Prometheus Prometheus
	Qr qr.Qr
}

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable  bool
	Address string
	Path    string

	Next    httpserver.Handler
	Handler http.Handler
}
