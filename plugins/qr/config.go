package qr

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/caddyserver/caddy"
	qrcode "github.com/skip2/go-qrcode"
)

// Qr contains Qr code generator's configuration
type Qr struct {
	Enable          bool
	Size            int
	BackgroundColor string
	ForegroundColor string
	RecoveryLevel   qrcode.RecoveryLevel

	BGColor color.Color
	FGColor color.Color
}

// ParseQr parses the config for QR requests
func (qr *Qr) ParseQr(c *caddy.Controller) error {
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
