package txtdirect

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"net/http"
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/skip2/go-qrcode"
)

// Qr contains Qr code generator's configuration
type Qr struct {
	Enable          bool
	Size            int
	BackgroundColor string
	ForegroundColor string
	RecoveryLevel   qrcode.RecoveryLevel

	backgroundColor color.Color
	foregroundColor color.Color
}

// Redirect handles the requests for "qr" requests
func (qr *Qr) Redirect(w http.ResponseWriter, r *http.Request) error {
	Qr, err := qrcode.New(r.Host+r.URL.String(), qr.RecoveryLevel)
	if err != nil {
		return fmt.Errorf("Couldn't generate the Qr instance: %s", err.Error())
	}
	if err = qr.ParseColors(); err != nil {
		return fmt.Errorf("Coudln't parse colors: %s", err.Error())
	}
	Qr.BackgroundColor, Qr.ForegroundColor = qr.backgroundColor, qr.foregroundColor

	Qr.Write(qr.Size, w)
	return nil
}

// ParseColors parses the hex colors in the QR config to color.Color instances
func (qr *Qr) ParseColors() error {
	bg, err := hex.DecodeString(qr.BackgroundColor)
	if err != nil {
		return err
	}
	qr.backgroundColor = color.RGBA{bg[0], bg[1], bg[2], bg[3]}
	fg, _ := hex.DecodeString(qr.ForegroundColor)
	if err != nil {
		return err
	}
	qr.foregroundColor = color.RGBA{fg[0], fg[1], fg[2], fg[3]}
	return nil
}

// ParseQr parses the config for QR requests
func (qr *Qr) ParseQr(c *caddyfile.Dispenser) error {
	switch c.Val() {
	case "size":
		value, err := strconv.Atoi(c.RemainingArgs()[0])
		if err != nil {
			return fmt.Errorf("<QR>: Couldn't parse the size")
		}
		qr.Size = value
	case "background":
		qr.BackgroundColor = c.RemainingArgs()[0]
	case "foreground":
		qr.ForegroundColor = c.RemainingArgs()[0]
	case "recovery_level":
		value, err := strconv.Atoi(c.RemainingArgs()[0])
		if err != nil {
			return fmt.Errorf("<QR>: Couldn't parse the recovery_level. It should be from 0 to 3")
		}
		qr.RecoveryLevel = qrcode.RecoveryLevel(value)
	default:
		return c.ArgErr() // unhandled option for QR config
	}
	return nil
}

// SetDefaults sets the default values for QR config
func (qr *Qr) SetDefaults() {
	if qr.Size == 0 {
		qr.Size = 256
	}
	if qr.BackgroundColor == "" {
		qr.BackgroundColor = "ffffffff"
	}
	if qr.ForegroundColor == "" {
		qr.ForegroundColor = "000000ff"
	}

	if len(qr.BackgroundColor) == 7 {
		qr.BackgroundColor = (qr.BackgroundColor + "ff")[1:]
	}
	if len(qr.ForegroundColor) == 7 {
		qr.ForegroundColor = (qr.ForegroundColor + "ff")[1:]
	}
}
