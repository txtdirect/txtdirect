/*
Copyright 2017 - The TXTdirect Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	From    string
	Default string
}

// Config contains the middleware's configuration
type Config struct {
	Enable   []string
	Redirect string
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

	if r.Type == "" {
		r.Type = "host"
	}

	return nil
}

func getBaseTarget(rec record) (string, int) {
	return rec.To, rec.Code
}

func getRecord(host, path string) (record, error) {
	if strings.Contains(host, ":") {
		hostSlice := strings.Split(host, ":")
		host = hostSlice[0]
	}
	zone := strings.Join([]string{basezone, host}, ".")

	var absoluteZone string
	if strings.HasSuffix(zone, ".") {
		absoluteZone = zone
	} else {
		absoluteZone = strings.Join([]string{zone, "."}, "")
	}
	s, err := net.LookupTXT(absoluteZone)

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

func contains(array []string, word string) bool {
	for _, w := range array {
		if w == word {
			return true
		}
	}
	return false
}

// Redirect the request depending on the redirect record found
func Redirect(w http.ResponseWriter, r *http.Request, c Config) error {
	host := r.Host
	path := r.URL.Path

	bl := make(map[string]bool)
	bl["/favicon.ico"] = true

	if bl[path] {
		http.Redirect(w, r, strings.Join([]string{host, path}, ""), 200)
		return nil
	}

	rec, err := getRecord(host, path)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			if c.Redirect != "" {
				http.Redirect(w, r, c.Redirect, http.StatusMovedPermanently)
				return nil
			}
			if contains(c.Enable, "www") {
				s := []string{defaultProtocol, "://", defaultSub, ".", host}
				http.Redirect(w, r, strings.Join(s, ""), 301)
				return nil
			}
			http.NotFound(w, r)
			return nil
		}
		return err
	}

	if !contains(c.Enable, rec.Type) {
		return fmt.Errorf("option disabled")
	}

	if path == "/" {
		http.Redirect(w, r, rec.Default, rec.Code)
		return nil
	}

	if rec.Type == "path" && path != "" {
		zone, from, _ := zoneFromPath(host, path, rec)
		rec, err = getFinalRecord(zone, from)
		if err != nil {
			return err
		}
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
