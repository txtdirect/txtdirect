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
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/plugins/prometheus"
	"go.txtdirect.org/txtdirect/record"
	"go.txtdirect.org/txtdirect/types"
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

func isIP(host string) bool {
	if v6slice := strings.Split(host, ":"); len(v6slice) > 2 {
		return true
	}
	hostSlice := strings.Split(host, ".")
	_, err := strconv.Atoi(hostSlice[len(hostSlice)-1])
	return err == nil
}

func blacklistRedirect(w http.ResponseWriter, r *http.Request, c config.Config) error {
	if bl[r.URL.Path] {
		redirect := strings.Join([]string{r.Host, r.URL.Path}, "")

		log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, redirect)
		// Empty Content-Type to prevent http.Redirect from writing an html response body
		w.Header().Set("Content-Type", "")
		w.Header().Add("Status-Code", strconv.Itoa(http.StatusNotFound))
		http.Redirect(w, r, redirect, http.StatusNotFound)
		if c.Prometheus.Enable {
			prometheus.RequestsByStatus.WithLabelValues(r.Host, strconv.Itoa(http.StatusNotFound)).Add(1)
		}
	}
	return nil
}

// Redirect the request depending on the redirect record found
func Redirect(w http.ResponseWriter, r *http.Request, c config.Config) error {
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
		record.Fallback(w, r, "global", http.StatusMovedPermanently, c)
		return nil
	}

	rec, err := record.GetRecord(host, c, w, r)
	r = rec.AddToContext(r)
	if err != nil {
		record.Fallback(w, r, "global", http.StatusFound, c)
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
		record.Fallback(w, r, "to", rec.Code, c)
		return nil
	}

	if rec.Type == "path" {
		prometheus.RequestsCountBasedOnType.WithLabelValues(host, "path").Add(1)
		prometheus.PathRedirectCount.WithLabelValues(host, path).Add(1)

		path := types.NewPath(w, r, path, rec, c)

		if path.Path == "/" {
			return path.RedirectRoot()
		}

		if path.Path != "" && rec.Re != "record" {
			record := path.Redirect()
			// It means fallback got triggered, If record is nil
			if record == nil {
				return nil
			}
			rec = *record
		}

		// Use predefined regexes if custom regex is set to "record"
		if path.Rec.Re == "record" {
			specificRec, err := path.SpecificRecord()
			if err != nil {
				log.Printf("[txtdirect]: Fallback is triggered because redirect to the most specific match failed: %s", err.Error())
				record.Fallback(path.Rw, path.Req, "to", path.Rec.Code, path.C)
				return nil
			}
			rec = *specificRec
		}
	}

	if rec.Type == "proxy" {
		prometheus.RequestsCountBasedOnType.WithLabelValues(host, "proxy").Add(1)
		log.Printf("[txtdirect]: %s > %s", rec.From, rec.To)

		proxy := types.NewProxy(w, r, rec, c)
		if err = proxy.Proxy(); err != nil {
			log.Print("Fallback is triggered because an error has occurred: ", err)
			record.Fallback(w, r, "to", rec.Code, c)
		}

		return nil
	}

	if rec.Type == "dockerv2" {
		prometheus.RequestsCountBasedOnType.WithLabelValues(host, "dockerv2").Add(1)

		docker := types.NewDockerv2(w, r, rec, c)

		if err := docker.Redirect(); err != nil {
			log.Printf("[txtdirect]: couldn't redirect to the requested container: %s", err.Error())
			record.Fallback(w, r, "to", rec.Code, c)
			return nil
		}
		return nil
	}

	if rec.Type == "host" {
		prometheus.RequestsCountBasedOnType.WithLabelValues(host, "host").Add(1)

		host := types.NewHost(w, r, rec, c)

		if err := host.Redirect(); err != nil {
			return err
		}
		return nil
	}

	if rec.Type == "gometa" {
		prometheus.RequestsCountBasedOnType.WithLabelValues(host, "gometa").Add(1)

		gometa := types.NewGometa(w, r, rec, c)

		// Triggers fallback when request isn't from `go get`
		if !gometa.ValidQuery() {
			return nil
		}

		return gometa.Serve()
	}

	if rec.Type == "git" {
		git := types.NewGit(w, r, c, rec)

		// Triggers fallback when request isn't from a Git client
		if !git.ValidGitQuery() {
			return nil
		}

		return git.Proxy()
	}

	return fmt.Errorf("record type %s unsupported", rec.Type)
}
