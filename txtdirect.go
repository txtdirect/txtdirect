package txtdirect

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const basezone = "_redirect"

type record struct {
	Version string
	To      string
	Code    int
}

func (r *record) Parse(str string) error {
	s := strings.Split(str, ";")
	for _, l := range s {
		switch {
		case strings.HasPrefix(l, "v="):
			l = strings.TrimPrefix(l, "v=")
			r.Version = l
			if r.Version != "txtv0" {
				return fmt.Errorf("unhandled version '%s'", r.Version)
			}
			log.Print("WARN: txtv0 is not suitable for production")

		case strings.HasPrefix(l, "to="):
			l = strings.TrimPrefix(l, "to=")
			r.To = l

		case strings.HasPrefix(l, "code="):
			l = strings.TrimPrefix(l, "code=")
			i, err := strconv.Atoi(l)
			if err != nil {
				return fmt.Errorf("could not parse status code: %s", err)
			}
			r.Code = i
		default:
			if r.To != "" {
				return fmt.Errorf("multiple values without keys")
			}
			r.To = l
		}
	}

	if r.Code == 0 {
		r.Code = 301
	}

	return nil
}

func getBaseTarget(host, path string) (string, int, error) {
	zone := basezone + "." + host
	s, err := net.LookupTXT(zone)
	if err != nil {
		return "", 0, err
	}

	rec := record{}
	if err = rec.Parse(s[0]); err != nil {
		return "", 0, err
	}

	if rec.To == "" {
		rec.To = "https://www." + host
	}

	return rec.To, rec.Code, nil
}

// Handle the redirection
func Handle(w http.ResponseWriter, r *http.Request) error {
	host := r.Host
	path := r.URL.Path

	to, code, err := getBaseTarget(host, path)
	if err != nil {
		return err
	}

	http.Redirect(w, r, to, code)
	return nil
}
