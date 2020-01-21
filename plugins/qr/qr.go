package qr

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"
)

// Redirect handles the requests for "qr" requests
func (qr *Qr) Redirect(w http.ResponseWriter, r *http.Request) error {
	Qr, err := qrcode.New(r.Host+r.URL.String(), qr.RecoveryLevel)
	if err != nil {
		return fmt.Errorf("Couldn't generate the Qr instance: %s", err.Error())
	}
	if err = qr.ParseColors(); err != nil {
		return fmt.Errorf("Coudln't parse colors: %s", err.Error())
	}
	Qr.BackgroundColor, Qr.ForegroundColor = qr.BGColor, qr.FGColor

	Qr.Write(qr.Size, w)
	return nil
}

// ParseColors parses the hex colors in the QR config to color.Color instances
func (qr *Qr) ParseColors() error {
	bg, err := hex.DecodeString(qr.BackgroundColor)
	if err != nil {
		return err
	}
	qr.BGColor = color.RGBA{bg[0], bg[1], bg[2], bg[3]}
	fg, _ := hex.DecodeString(qr.ForegroundColor)
	if err != nil {
		return err
	}
	qr.FGColor = color.RGBA{fg[0], fg[1], fg[2], fg[3]}
	return nil
}
