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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/caddy/caddyhttp/proxy"
)

const (
	basezone        = "_redirect"
	defaultSub      = "www"
	defaultProtocol = "https"
	proxyKeepalive  = 30
	logFormat       = "02/Jan/2006:15:04:05 -0700"
	proxyTimeout    = 30 * time.Second
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
	Enable     []string
	Redirect   string
	Resolver   string
	Prometheus Prometheus
}

func (r *record) Parse(str string, req *http.Request) error {
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
			l, err := parsePlaceholders(l, req)
			if err != nil {
				return err
			}
			r.From = l

		case strings.HasPrefix(l, "re="):
			l = strings.TrimPrefix(l, "re=")
			r.Re = l

		case strings.HasPrefix(l, "root="):
			l = strings.TrimPrefix(l, "root=")
			r.Root = l

		case strings.HasPrefix(l, "to="):
			l = strings.TrimPrefix(l, "to=")
			l, err := parsePlaceholders(l, req)
			if err != nil {
				return err
			}
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

func getBaseTarget(rec record, r *http.Request) (string, int, error) {
	if strings.ContainsAny(rec.To, "{}") {
		to, err := parsePlaceholders(rec.To, r)
		if err != nil {
			return "", 0, err
		}
		rec.To = to
	}
	return rec.To, rec.Code, nil
}

func contains(array []string, word string) bool {
	for _, w := range array {
		if w == word {
			return true
		}
	}
	return false
}

func getRecord(host, path string, ctx context.Context, c Config, r *http.Request) (record, error) {
	txts, err := query(host, ctx, c)
	if err != nil {
		return record{}, err
	}

	if len(txts) != 1 {
		return record{}, fmt.Errorf("could not parse TXT record with %d records", len(txts))
	}

	rec := record{}
	if err = rec.Parse(txts[0], r); err != nil {
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
	} else {
		http.NotFound(w, r)
	}
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

	rec, err := getRecord(host, path, r.Context(), c, r)
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

	fallbackURL, code, err := getBaseTarget(rec, r)
	if err != nil {
		return err
	}

	if rec.Re != "" && rec.From != "" {
		fallback(w, r, fallbackURL, code, c)
		return nil
	}

	if rec.Type == "path" && contains(c.Enable, rec.Type) {
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
			rec, err = getFinalRecord(zone, from, r.Context(), c, r)
			if err != nil {
				log.Print("Fallback is triggered because an error has occurred: ", err)
				fallback(w, r, fallbackURL, code, c)
				return nil
			}
		}
	}

	if rec.Type == "proxy" && contains(c.Enable, rec.Type) {
		log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), rec.From, rec.To)
		to, _, err := getBaseTarget(rec, r)
		if err != nil {
			log.Print("Fallback is triggered because an error has occurred: ", err)
			fallback(w, r, fallbackURL, code, c)
			return err
		}
		u, err := url.Parse(to)
		if err != nil {
			return err
		}
		reverseProxy := proxy.NewSingleHostReverseProxy(u, "", proxyKeepalive, proxyTimeout)
		reverseProxy.ServeHTTP(w, r, nil)
		return nil
	}

	if rec.Type == "host" && contains(c.Enable, rec.Type) {
		to, code, err := getBaseTarget(rec, r)
		if err != nil {
			log.Print("Fallback is triggered because an error has occurred: ", err)
			fallback(w, r, fallbackURL, code, c)
			return err
		}
		log.Printf("<%s> [txtdirect]: %s > %s", time.Now().Format(logFormat), r.URL.String(), to)
		http.Redirect(w, r, to, code)
		return nil
	}

	if rec.Type == "gometa" && contains(c.Enable, rec.Type) {
		return gometa(w, rec, host, path)
	}

	return fmt.Errorf("record type %s unsupported", rec.Type)
}
