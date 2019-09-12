package txtdirect

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy/caddyhttp/proxy"
	"github.com/prometheus/common/log"
)

type Git struct {
	rw  http.ResponseWriter
	req *http.Request
	c   Config
	rec record
}

func NewGit(w http.ResponseWriter, r *http.Request, c Config, rec record) *Git {
	return &Git{
		rw:  w,
		req: r,
		c:   c,
		rec: rec,
	}
}

func (g *Git) Proxy() error {
	to, _, err := getBaseTarget(g.rec, g.req)
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

	reverseProxy := proxy.NewSingleHostReverseProxy(u, "", proxyKeepalive, proxyTimeout, fallbackDelay)

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

func (g *Git) ValidGitQuery() bool {
	if !strings.HasPrefix(g.req.Header.Get("User-Agent"), "git") {
		fallback(g.rw, g.req, "website", g.rec.Code, g.c)
		log.Errorf("[txtdirect]: The incoming request is not from a Git client.")
		return false
	}
	return true
}
