package caddy

import (
	"net/http"

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

func setup(c *caddy.Controller) error {
	var enable, disable []string
	c.Next() // skip directive name
	for c.NextBlock() {
		option := c.Val()
		switch option {
		case "enable":
			if disable != nil {
				return c.ArgErr()
			}
			enable = c.RemainingArgs()
		case "disable":
			if enable != nil {
				return c.ArgErr()
			}
			disable = c.RemainingArgs()
		default:
			return c.ArgErr() // unhandled option
		}
	}

	// Add handler to Caddy
	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return Redirect{
			Next:    next,
			Enable:  enable,
			Disable: disable,
		}
	}
	cfg.AddMiddleware(mid)
	return nil
}

// Redirect is middleware to redirect requests based on TXT records
type Redirect struct {
	Next    httpserver.Handler
	Enable  []string
	Disable []string
}

func (rd Redirect) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if err := txtdirect.Redirect(w, r); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
