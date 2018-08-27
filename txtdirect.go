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
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	basezone        = "_redirect"
	defaultSub      = "www"
	defaultProtocol = "https"
	logFormat       = "02/Jan/2006:15:04:05 -0700"
)

var PlaceholderRegex = regexp.MustCompile("{[~>?]?\\w+}")

type record struct {
	Version string
	To      string
	Code    int
	Type    string
	Vcs     string
	From    string
	Root    string
	Re      string
}

// Config contains the middleware's configuration
type Config struct {
	Enable   []string
	Redirect string
	Resolver string
}

func (r *record) Parse(str string) error {
	s := strings.Split(str, ";")
	for _, l := range s {
		switch {
		case strings.HasPrefix(l, "code="):
			l = strings.TrimPrefix(l, "code=")
			i, err := strconv.Atoi(l)
			if err != nil {
				return fmt.Errorf("could not parse status code: %s", err)
			}
			r.Code = i

		case strings.HasPrefix(l, "from="):
			l = strings.TrimPrefix(l, "from=")
			r.From = l

		case strings.HasPrefix(l, "re="):
			l = strings.TrimPrefix(l, "re=")
			r.Re = l

		case strings.HasPrefix(l, "root="):
			l = strings.TrimPrefix(l, "root=")
			r.Root = l

		case strings.HasPrefix(l, "to="):
			l = strings.TrimPrefix(l, "to=")
			r.To = l

		case strings.HasPrefix(l, "type="):
			l = strings.TrimPrefix(l, "type=")
			r.Type = l

		case strings.HasPrefix(l, "v="):
			l = strings.TrimPrefix(l, "v=")
			r.Version = l
			if r.Version != "txtv0" {
				return fmt.Errorf("unhandled version '%s'", r.Version)
			}
			log.Print("WARN: txtv0 is not suitable for production")

		case strings.HasPrefix(l, "vcs="):
			l = strings.TrimPrefix(l, "vcs=")
			r.Vcs = l

		default:
			tuple := strings.Split(l, "=")
			if len(tuple) != 2 {
				return fmt.Errorf("arbitrary data not allowed")
			}
			continue
		}
		if len(l) > 255 {
			return fmt.Errorf("TXT record cannot exceed the maximum of 255 characters")
		}
	}

	if r.Code == 0 {
		r.Code = 301
	}

	if r.Vcs == "" && r.Type == "gometa" {
		r.Vcs = "git"
	}

	if r.Type == "" {
		r.Type = "host"
	}

	return nil
}

func getBaseTarget(rec record, r *http.Request) (string, int) {
	if strings.ContainsAny(rec.To, "{}") {
		rec.To = parsePlaceholders(rec.To, r)
	}
	return rec.To, rec.Code
}

func parsePlaceholders(input string, r *http.Request) string {
	placeholders := PlaceholderRegex.FindAllStringSubmatch(input, -1)
	for _, placeholder := range placeholders {
		switch placeholder[0] {
		case "{uri}":
			input = strings.Replace(input, "{uri}", r.URL.RequestURI(), -1)
		case "{dir}":
			dir, _ := path.Split(r.URL.Path)
			input = strings.Replace(input, "{dir}", dir, -1)
		case "{file}":
			_, file := path.Split(r.URL.Path)
			input = strings.Replace(input, "{file}", file, -1)
		case "{fragment}":
			input = strings.Replace(input, "{fragment}", r.URL.Fragment, -1)
		case "{host}":
			input = strings.Replace(input, "{host}", r.URL.Host, -1)
		case "{hostonly}":
			input = strings.Replace(input, "{hostonly}", r.URL.Hostname(), -1)
		case "{method}":
			input = strings.Replace(input, "{method}", r.Method, -1)
		case "{path}":
			input = strings.Replace(input, "{path}", r.URL.Path, -1)
		case "{path_escaped}":
			input = strings.Replace(input, "{path_escaped}", url.QueryEscape(r.URL.Path), -1)
		case "{port}":
			input = strings.Replace(input, "{port}", r.URL.Port(), -1)
		case "{query}":
			input = strings.Replace(input, "{query}", r.URL.RawQuery, -1)
		case "{query_escaped}":
			input = strings.Replace(input, "{query_escaped}", url.QueryEscape(r.URL.RawQuery), -1)
		case "{uri_escaped}":
			input = strings.Replace(input, "{uri_escaped}", url.QueryEscape(r.URL.RequestURI()), -1)
		case "{user}":
			user, _, ok := r.BasicAuth()
			if !ok {
				input = strings.Replace(input, "{user}", "", -1)
			}
			input = strings.Replace(input, "{user}", user, -1)
		}
		if placeholder[0][1] == '>' {
			want := placeholder[0][2 : len(placeholder[0])-1]
			for key, values := range r.Header {
				// Header placeholders (case-insensitive)
				if strings.EqualFold(key, want) {
					input = strings.Replace(input, placeholder[0], strings.Join(values, ","), -1)
				}
			}
		}
		if placeholder[0][1] == '~' {
			name := placeholder[0][2 : len(placeholder[0])-1]
			if cookie, err := r.Cookie(name); err == nil {
				input = strings.Replace(input, placeholder[0], cookie.Value, -1)
			}
		}
		if placeholder[0][1] == '?' {
			query := r.URL.Query()
			name := placeholder[0][2 : len(placeholder[0])-1]
			input = strings.Replace(input, placeholder[0], query.Get(name), -1)
		}
	}
	return input
}

func contains(array []string, word string) bool {
	for _, w := range array {
		if w == word {
			return true
		}
	}
	return false
}

func getRecord(host, path string, ctx context.Context, c Config) (record, error) {
	txts, err := query(host, ctx, c)
	if err != nil {
		return record{}, err
	}

	if len(txts) != 1 {
		return record{}, fmt.Errorf("could not parse TXT record with %d records", len(txts))
	}

	rec := record{}
	if err = rec.Parse(txts[0]); err != nil {
		return rec, fmt.Errorf("could not parse record: %s", err)
	}

	return rec, nil
}

func fallback(w http.ResponseWriter, r *http.Request, fallback string, code int, c Config) {
	if fallback != "" {
		log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), fallback)
		http.Redirect(w, r, fallback, code)
	} else if c.Redirect != "" {
		for _, enable := range c.Enable {
			if enable == "www" {
				log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), c.Redirect)
				http.Redirect(w, r, c.Redirect, 403)
			}
		}
	}
	http.NotFound(w, r)
}

func customResolver(c Config) net.Resolver {
	return net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, c.Resolver)
		},
	}
}

func query(zone string, ctx context.Context, c Config) ([]string, error) {
	// Removes port from zone
	if strings.Contains(zone, ":") {
		zoneSlice := strings.Split(zone, ":")
		zone = zoneSlice[0]
	}

	if !strings.HasPrefix(zone, basezone) {
		zone = strings.Join([]string{basezone, zone}, ".")
	}

	// Use absolute zone
	var absoluteZone string
	if strings.HasSuffix(zone, ".") {
		absoluteZone = zone
	} else {
		absoluteZone = strings.Join([]string{zone, "."}, "")
	}

	var txts []string
	var err error
	if c.Resolver != "" {
		net := customResolver(c)
		txts, err = net.LookupTXT(ctx, absoluteZone)
	} else {
		txts, err = net.LookupTXT(absoluteZone)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get TXT record: %s", err)
	}
	return txts, nil
}

// Redirect the request depending on the redirect record found
func Redirect(w http.ResponseWriter, r *http.Request, c Config) error {
	host := r.Host
	path := r.URL.Path

	bl := make(map[string]bool)
	bl["/favicon.ico"] = true

	if bl[path] {
		redirect := strings.Join([]string{host, path}, "")
		log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), redirect)
		http.Redirect(w, r, redirect, 200)
		return nil
	}

	rec, err := getRecord(host, path, r.Context(), c)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			if c.Redirect != "" {
				log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), c.Redirect)
				http.Redirect(w, r, c.Redirect, http.StatusMovedPermanently)
				return nil
			}
			if contains(c.Enable, "www") {
				s := strings.Join([]string{defaultProtocol, "://", defaultSub, ".", host}, "")
				log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), s)
				http.Redirect(w, r, s, 301)
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

	fallbackURL, code := getBaseTarget(rec, r)

	if rec.Re != "" && rec.From != "" {
		fallback(w, r, fallbackURL, code, c)
	}

	if rec.Type == "path" {
		if path == "/" {
			if rec.Root == "" {
				fallback(w, r, fallbackURL, code, c)
				return nil
			}
			log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), rec.Root)
			http.Redirect(w, r, rec.Root, rec.Code)
			return nil
		}

		if path != "" {
			zone, from, err := zoneFromPath(host, path, rec)
			rec, err = getFinalRecord(zone, from, r.Context(), c)
			if err != nil {
				log.Print("Fallback is triggered because an error has occurred: ", err)
				fallback(w, r, fallbackURL, code, c)
				return err
			}
		}
	}

	if rec.Type == "host" {
		to, code := getBaseTarget(rec, r)
		log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), to)
		http.Redirect(w, r, to, code)
		return nil
	}

	if rec.Type == "gometa" {
		return gometa(w, rec, host, path)
	}

	return fmt.Errorf("record type %s unsupported", rec.Type)
}
