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

var allOptions = []string{"host", "gometa"}

func setup(c *caddy.Controller) error {
	var enable []string
	c.Next() // skip directive name
	for c.NextBlock() {
		option := c.Val()
		switch option {
		case "enable":
			if enable != nil {
				return c.ArgErr()
			}
			enable = c.RemainingArgs()
		case "disable":
			if enable != nil {
				return c.ArgErr()
			}
			enable = removeArrayFromArray(enable, c.RemainingArgs())
		default:
			return c.ArgErr() // unhandled option
		}
	}

	// If nothing is specified, enable everything
	if enable == nil {
		enable = allOptions
	}

	// Add handler to Caddy
	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return Redirect{
			Next:   next,
			Enable: enable,
		}
	}
	cfg.AddMiddleware(mid)
	return nil
}

func removeArrayFromArray(array, toBeRemoved []string) []string {
	for _, toRemove := range toBeRemoved {
		for i, option := range array {
			if option == toRemove {
				array[i] = array[len(array)-1]
				array = array[:len(array)-1]
				break
			}
		}
	}
	return array
}

// Redirect is middleware to redirect requests based on TXT records
type Redirect struct {
	Next   httpserver.Handler
	Enable []string
}

func (rd Redirect) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if err := txtdirect.Redirect(w, r, rd.Enable); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
