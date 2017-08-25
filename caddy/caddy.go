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
	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return Handler{Next: next}
	}
	cfg.AddMiddleware(mid)
	return nil
}

type Handler struct {
	Next httpserver.Handler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if err := txtdirect.Handle(w, r); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
