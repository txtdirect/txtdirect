package config

import (
	"go.txtdirect.org/txtdirect/plugins/prometheus"
	"go.txtdirect.org/txtdirect/plugins/qr"
)

// Config contains the middleware's configuration
type Config struct {
	Enable     []string
	Redirect   string
	Resolver   string
	LogOutput  string
	Prometheus prometheus.Prometheus
	Qr         qr.Qr
}
