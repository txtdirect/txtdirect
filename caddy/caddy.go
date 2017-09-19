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

func parse(c *caddy.Controller) ([]string, error) {
	var enable []string
	c.Next() // skip directive name
	for c.NextBlock() {
		option := c.Val()
		switch option {
		case "enable":
			if enable != nil {
				return enable, c.ArgErr()
			}
			enable = c.RemainingArgs()
			if len(enable) == 0 {
				return enable, c.ArgErr()
			}

		case "disable":
			if enable != nil {
				return enable, c.ArgErr()
			}
			toDisable := c.RemainingArgs()
			if len(toDisable) == 0 {
				return enable, c.ArgErr()
			}
			enable = removeArrayFromArray(allOptions, toDisable)

		default:
			return enable, c.ArgErr() // unhandled option
		}
	}

	// If nothing is specified, enable everything
	if enable == nil {
		enable = allOptions
	}

	return enable, nil
}

func setup(c *caddy.Controller) error {
	enable, err := parse(c)
	if err != nil {
		return err
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
		if err.Error() == "option disabled" {
			return rd.Next.ServeHTTP(w, r)
		}
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
