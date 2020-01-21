package config

import (
	"image/color"
	"net/http"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	qrcode "github.com/skip2/go-qrcode"
)

// Config contains the middleware's configuration
type Config struct {
	Enable    []string
	Redirect  string
	Resolver  string
	LogOutput string
	Prometheus Prometheus
	Qr Qr
}

type Qr struct {
	Enable          bool
	Size            int
	BackgroundColor string
	ForegroundColor string
	RecoveryLevel   qrcode.RecoveryLevel

	BGColor color.Color
	FGColor color.Color
}

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable  bool
	Address string
	Path    string

	Next    httpserver.Handler
	Handler http.Handler
}
