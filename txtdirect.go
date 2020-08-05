/*
Copyright 2017 - The TXTDirect Authors
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
	"strconv"
	"strings"
	"time"
)

const (
	basezone          = "_redirect"
	defaultSub        = "www"
	defaultProtocol   = "https"
	proxyKeepalive    = 30
	fallbackDelay     = 300 * time.Millisecond
	proxyTimeout      = 30 * time.Second
	Status301CacheAge = 604800
)

var bl = map[string]bool{
	"/favicon.ico": true,
}

// Redirect the request depending on the redirect record found
func Redirect(w http.ResponseWriter, r *http.Request, c Config) error {
	w.Header().Set("Server", "TXTDirect")

	host := r.Host
	path := r.URL.Path

	if c.Qr.Enable {
		// Return the Qr code for the URI if "qr" query is available
		if _, ok := r.URL.Query()["qr"]; ok {
			return c.Qr.Redirect(w, r)
		}
	}

	// Check the blacklist and redirect to the request's host and path
	if err := blacklistRedirect(w, r, c); err != nil {
		return err
	}

	if isIP(host) {
		log.Println("[txtdirect]: Trying to access 127.0.0.1, fallback triggered.")
		fallback(w, r, "global", http.StatusMovedPermanently, c)
		return nil
	}

	rec, err := GetRecord(host, c, w, r)
	if err != nil {
		fallback(w, r, "global", http.StatusFound, c)
		return nil
	}

	// Add the upstream zone address from the use= fields to the request context
	if r, err = rec.CheckUpstream(w, r, c); err != nil {
		log.Printf("[txtdirect]: Couldn't fetch the upstream record: %s", err.Error())
		fallback(w, r, "global", http.StatusFound, c)
		return nil
	}

	r = rec.addToContext(r)

	// Add referer header
	if rec.Ref && r.Header.Get("Referer") == "" {
		host := r.Host
		if strings.Contains(host, ":") {
			hostSlice := strings.Split(host, ":")
			host = hostSlice[0]
		}
		w.Header().Set("Referer", host)
	}

	if !contains(c.Enable, rec.Type) {
		return fmt.Errorf("type \"%s\" is not enabled. Enabled types are: %v", rec.Type, c.Enable)
	}

	if rec.Re != "" && rec.From != "" {
		log.Println("[txtdirect]: It's not allowed to use both re= and from= in a record.")
		fallback(w, r, "to", rec.Code, c)
		return nil
	}

	if rec.Type == "path" {
		path := NewPath(w, r, path, rec, c)

		if path.path == "/" {
			return path.RedirectRoot()
		}

		if path.path != "" && rec.Re != "record" {
			record := path.Redirect()
			// It means fallback got triggered, If record is nil
			if record == nil {
				return nil
			}
			rec = *record
		}

		// Use predefined regexes if custom regex is set to "record"
		if path.rec.Re == "record" {
			record, err := path.SpecificRecord()
			if err != nil {
				log.Printf("[txtdirect]: Fallback is triggered because redirect to the most specific match failed: %s", err.Error())
				fallback(path.rw, path.req, "to", path.rec.Code, path.c)
				return nil
			}
			rec = *record
		}
	}

	if rec.Type == "host" {
		host := NewHost(w, r, rec, c)

		if err := host.Redirect(); err != nil {
			return err
		}
		return nil
	}

	if rec.Type == "gometa" {
		gometa := NewGometa(w, r, rec, c)

		// Triggers fallback when request isn't from `go get`
		if !gometa.ValidQuery() {
			return nil
		}

		return gometa.Serve()
	}

	return fmt.Errorf("record type %s unsupported", rec.Type)
}

// UpstreamZone returns the upstream zone from request's context
func UpstreamZone(r *http.Request) string {
	if zone := r.Context().Value("upstreamZone"); zone != nil {
		return zone.(string)
	}
	return r.Host
}

// customResolver returns a net.Resolver instance based
// on the given txtdirect config to use a custom DNS resolver.
func customResolver(c Config) net.Resolver {
	return net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, c.Resolver)
		},
	}
}

func isIP(host string) bool {
	if v6slice := strings.Split(host, ":"); len(v6slice) > 2 {
		return true
	}
	hostSlice := strings.Split(host, ".")
	_, err := strconv.Atoi(hostSlice[len(hostSlice)-1])
	return err == nil
}

func blacklistRedirect(w http.ResponseWriter, r *http.Request, c Config) error {
	if bl[r.URL.Path] {
		redirect := strings.Join([]string{r.Host, r.URL.Path}, "")

		log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, redirect)
		// Empty Content-Type to prevent http.Redirect from writing an html response body
		w.Header().Set("Content-Type", "")
		w.Header().Add("Status-Code", strconv.Itoa(http.StatusNotFound))
		http.Redirect(w, r, redirect, http.StatusNotFound)
	}
	return nil
}

// contains checks the given slice to see if an item exists
// in that slice or not
func contains(array []string, word string) bool {
	for _, w := range array {
		if w == word {
			return true
		}
	}
	return false
}
