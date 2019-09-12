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
	status301CacheAge = 604800
)

var bl = map[string]bool{
	"/favicon.ico": true,
}

// Config contains the middleware's configuration
type Config struct {
	Enable     []string
	Redirect   string
	Resolver   string
	LogOutput  string
	Prometheus Prometheus
	Qr         Qr
}

// getBaseTarget parses the placeholder in the given record's To= field
// and returns the final address and http status code
func getBaseTarget(rec record, r *http.Request) (string, int, error) {
	if strings.ContainsAny(rec.To, "{}") {
		to, err := parsePlaceholders(rec.To, r, []string{})
		if err != nil {
			return "", 0, err
		}
		rec.To = to
	}
	return rec.To, rec.Code, nil
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

// query checks the given zone using net.LookupTXT to
// find TXT records in that zone
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
		if c.Prometheus.Enable {
			RequestsByStatus.WithLabelValues(r.Host, strconv.Itoa(http.StatusNotFound)).Add(1)
		}
	}
	return nil
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

	rec, err := getRecord(host, c, w, r)
	r = rec.addToContext(r)
	if err != nil {
		fallback(w, r, "global", http.StatusFound, c)
		return nil
	}

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
		RequestsCountBasedOnType.WithLabelValues(host, "path").Add(1)
		PathRedirectCount.WithLabelValues(host, path).Add(1)

		path := NewPath(w, r, path, rec, c)

		if path.path == "/" {
			return path.RedirectRoot()
		}

		if path.path != "" {
			record := path.Redirect()
			// It means fallback got triggered, If record is nil
			if record == nil {
				return nil
			}
			rec = *record
		}
	}

	if rec.Type == "proxy" {
		RequestsCountBasedOnType.WithLabelValues(host, "proxy").Add(1)
		log.Printf("[txtdirect]: %s > %s", rec.From, rec.To)

		proxy := NewProxy(w, r, rec, c)
		if err = proxy.Proxy(); err != nil {
			log.Print("Fallback is triggered because an error has occurred: ", err)
			fallback(w, r, "to", rec.Code, c)
		}

		return nil
	}

	if rec.Type == "dockerv2" {
		RequestsCountBasedOnType.WithLabelValues(host, "dockerv2").Add(1)

		docker := NewDockerv2(w, r, rec, c)

		// Fallback gets triggered if the User-Agent isn't valid
		if !docker.ValidAgent() {
			return nil
		}

		if err := docker.Redirect(); err != nil {
			log.Printf("[txtdirect]: couldn't redirect to the requested container: %s", err.Error())
			fallback(w, r, "to", rec.Code, c)
			return nil
		}
		return nil
	}

	if rec.Type == "host" {
		RequestsCountBasedOnType.WithLabelValues(host, "host").Add(1)

		host := NewHost(w, r, rec, c)

		if err := host.Redirect(); err != nil {
			return err
		}
		return nil
	}

	if rec.Type == "gometa" {
		RequestsCountBasedOnType.WithLabelValues(host, "gometa").Add(1)

		gometa := NewGometa(w, r, rec, c)

		// Triggers fallback when request isn't from `go get`
		if !gometa.ValidQuery() {
			return nil
		}

		return gometa.Serve()
	}

	if rec.Type == "git" {
		git := NewGit(w, r, c, rec)

		// Triggers fallback when request isn't from a Git client
		if !git.ValidGitQuery() {
			return nil
		}

		return git.Proxy()
	}

	return fmt.Errorf("record type %s unsupported", rec.Type)
}
