package txtdirect

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	basezone        = "_redirect"
	defaultSub      = "www"
	defaultProtocol = "https"
)

type record struct {
	Version string
	To      string
	Code    int
	Type    string
	Vcs     string
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

		case strings.HasPrefix(l, "type="):
			l = strings.TrimPrefix(l, "type=")
			r.Type = l

		case strings.HasPrefix(l, "vcs="):
			l = strings.TrimPrefix(l, "vcs=")
			r.Vcs = l

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

	if r.Vcs == "" {
		r.Vcs = "git"
	}

	if r.Type == "" {
		r.Type = "host"
	}

	return nil
}

func getBaseTarget(rec record) (string, int) {
	return rec.To, rec.Code
}

func getRecord(host, path string) (record, error) {
	zone := strings.Join([]string{basezone, host}, ".")
	s, err := net.LookupTXT(zone)
	if err != nil {
		return record{}, fmt.Errorf("could not get TXT record: %s", err)
	}

	rec := record{}
	if err = rec.Parse(s[0]); err != nil {
		return rec, fmt.Errorf("could not parse record: %s", err)
	}

	if rec.To == "" {
		s := []string{defaultProtocol, "://", defaultSub, ".", host}
		rec.To = strings.Join(s, "")
	}

	return rec, nil
}

// Redirect the request depending on the redirect record found
func Redirect(w http.ResponseWriter, r *http.Request) error {
	host := r.Host
	path := r.URL.Path

	rec, err := getRecord(host, path)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			s := []string{defaultProtocol, "://", defaultSub, ".", host}
			http.Redirect(w, r, strings.Join(s, ""), 301)
			return nil
		}
		return err
	}

	if rec.Type == "host" {
		to, code := getBaseTarget(rec)
		http.Redirect(w, r, to, code)
		return nil
	}

	if rec.Type == "gometa" {
		return gometa(w, rec, host, path)
	}

	return fmt.Errorf("record type %s unsupported", rec.Type)
}
