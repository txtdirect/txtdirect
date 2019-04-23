package main

import (
	"github.com/mholt/caddy/caddy/caddymain"

	_ "github.com/SchumacherFM/mailout"
	_ "github.com/captncraig/caddy-realip"
	_ "github.com/miekg/caddy-prometheus"
	_ "github.com/txtdirect/txtdirect/caddy"
)

func main() {
	caddymain.EnableTelemetry = false
	caddymain.Run()
}
