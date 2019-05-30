package txtdirect

import (
	"strconv"

	"github.com/mholt/caddy"
)

// Qr contains Qr code generator's configuration
type Qr struct {
	Size            int
	BackgroundColor string
	ForegroundColor string
}

func (qr *Qr) ParseQr(c *caddy.Controller) error {
	switch c.Val() {
	case "size":
		value, err := strconv.Atoi(c.RemainingArgs()[0])
		if err != nil {
			return c.ArgErr()
		}
		qr.Size = value
	case "background":
		qr.BackgroundColor = c.RemainingArgs()[0]
	case "foreground":
		qr.ForegroundColor = c.RemainingArgs()[0]
	default:
		return c.ArgErr() // unhandled option for QR config
	}
	return nil
}

func (qr *Qr) SetDefaults() {
	if qr.Size == 0 {
		qr.Size = 256
	}
	if qr.BackgroundColor == "" {
		qr.BackgroundColor = "#ffffff"
	}
	if qr.ForegroundColor == "" {
		qr.ForegroundColor = "#000000"
	}
}
