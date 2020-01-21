package types

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy/caddyhttp/proxy"
	"github.com/prometheus/common/log"
	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/record"
	"go.txtdirect.org/txtdirect/variables"
)

// Git keeps data for "git" type requests
type Git struct {
	rw  http.ResponseWriter
	req *http.Request
	c   config.Config
	rec record.Record
}

// NewGit returns a fresh instance of Git struct
func NewGit(w http.ResponseWriter, r *http.Request, c config.Config, rec record.Record) *Git {
	return &Git{
		rw:  w,
		req: r,
		c:   c,
		rec: rec,
	}
}

// Proxy handles the requests for "git" type
func (g *Git) Proxy() error {
	to, _, err := record.GetBaseTarget(g.rec, g.req)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(g.req.URL.Path, "/info") {
		g.req.URL.Path = "/" + strings.Join(strings.Split(g.req.URL.Path, "/")[2:], "/")
	}

	u, err := url.Parse(to)
	if err != nil {
		return err
	}

	reverseProxy := proxy.NewSingleHostReverseProxy(u, "", variables.ProxyKeepalive, variables.ProxyTimeout, variables.FallbackDelay)

	tmpResponse := ProxyResponse{headers: make(http.Header)}
	reverseProxy.ServeHTTP(&tmpResponse, g.req, nil)

	// Decompress the body based on "Content-Encoding" header and write to a writer buffer
	if err := tmpResponse.WriteBody(); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the response body: %s", err.Error())
	}

	// Replace the URL hosts with the request's host
	if err := tmpResponse.ReplaceBody(u.Scheme, u.Host, g.req.Host); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't replace urls inside the response body: %s", err.Error())
	}

	copyHeader(g.rw.Header(), tmpResponse.Header())

	// Write the status from the temporary ResponseWriter to the main ResponseWriter
	g.rw.WriteHeader(tmpResponse.status)

	// Write the final response from the temporary ResponseWriter to the main ResponseWriter
	if _, err := g.rw.Write(tmpResponse.Body()); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the temporary response to main response body: %s", err.Error())
	}
	return nil
}

// ValidGitQuery checks the User-Agent header to make sure the requests are
// coming from a Git client.
func (g *Git) ValidGitQuery() bool {
	if !strings.HasPrefix(g.req.Header.Get("User-Agent"), "git") {
		record.Fallback(g.rw, g.req, "website", g.rec.Code, g.c)
		log.Errorf("[txtdirect]: The incoming request is not from a Git client.")
		return false
	}
	return true
}
