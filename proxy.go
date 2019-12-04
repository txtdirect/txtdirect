/*
Copyright 2019 - The TXTDirect Authors
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
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/caddyserver/caddy/caddyhttp/proxy"
)

// ProxyResponse keeps the custom response writer data and is
// a custom implementation of net/http's default response writer
type ProxyResponse struct {
	headers    http.Header
	body       []byte
	bodyReader bytes.Buffer
	bodyWriter bytes.Buffer
	status     int
}

// Proxy contains the data that are needed to proxy the request
type Proxy struct {
	rw  http.ResponseWriter
	req *http.Request
	c   Config
	rec record
}

// NewProxy returns a fresh instance of Proxy struct
func NewProxy(w http.ResponseWriter, r *http.Request, rec record, c Config) *Proxy {
	return &Proxy{
		rw:  w,
		req: r,
		rec: rec,
		c:   c,
	}
}

// Proxy proxies the request to the endpoint defined in the record
func (p *Proxy) Proxy() error {
	to, _, err := getBaseTarget(p.rec, p.req)
	if err != nil {
		return err
	}
	u, err := url.Parse(to)
	if err != nil {
		return err
	}
	reverseProxy := proxy.NewSingleHostReverseProxy(u, "", proxyKeepalive, proxyTimeout, fallbackDelay)

	tmpResponse := ProxyResponse{headers: make(http.Header)}
	reverseProxy.ServeHTTP(&tmpResponse, p.req, nil)

	// Decompress the body based on "Content-Encoding" header and write to a writer buffer
	if err := tmpResponse.WriteBody(); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the response body: %s", err.Error())
	}

	// Replace the URL hosts with the request's host
	if err := tmpResponse.ReplaceBody(u.Scheme, u.Host, p.req.Host); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't replace urls inside the response body: %s", err.Error())
	}

	copyHeader(p.rw.Header(), tmpResponse.Header())

	// Write the status from the temporary ResponseWriter to the main ResponseWriter
	p.rw.WriteHeader(tmpResponse.status)

	// Write the final response from the temporary ResponseWriter to the main ResponseWriter
	if _, err := p.rw.Write(tmpResponse.Body()); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the temporary response to main response body: %s", err.Error())
	}
	return nil
}

// Header returns response headers
func (p *ProxyResponse) Header() http.Header {
	return p.headers
}

func (p *ProxyResponse) Write(body []byte) (int, error) {
	reader := bytes.NewReader(body)
	pooledIoCopy(&p.bodyReader, reader)
	p.body = body
	return len(body), nil
}

// Body returns response's body. This method should only get called after WriteBody()
func (p *ProxyResponse) Body() []byte {
	return p.bodyWriter.Bytes()
}

// WriteHeader Writes the given status code to response
func (p *ProxyResponse) WriteHeader(status int) {
	p.status = status
}

// ReplaceBody replaces the URIs inside the response body
// to use the request's host instead of the endpoint's host
func (p *ProxyResponse) ReplaceBody(scheme, to, host string) error {
	replacedBody := bytes.Replace(p.bodyWriter.Bytes(), []byte(scheme+"://"+to), []byte(scheme+"://"+host), -1)
	p.bodyWriter.Reset()
	if _, err := p.bodyWriter.Write(replacedBody); err != nil {
		return err
	}
	return nil
}

// WriteBody writes the response body to custom response writer's body
func (p *ProxyResponse) WriteBody() error {
	switch p.Header().Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(&p.bodyReader)
		if err != nil {
			return err
		}
		defer reader.Close()
		_, err = io.Copy(&p.bodyWriter, reader)
		if err != nil {
			return err
		}
		p.Header().Del("Content-Encoding")
	default:
		_, err := io.Copy(&p.bodyWriter, &p.bodyReader)
		if err != nil {
			return err
		}
	}
	return nil
}

var bufferPool = sync.Pool{New: createBuffer}

func createBuffer() interface{} {
	return make([]byte, 0, 32*1024)
}

var skipHeaders = map[string]struct{}{
	"Content-Type":        {},
	"Content-Disposition": {},
	"Accept-Ranges":       {},
	"Set-Cookie":          {},
	"Cache-Control":       {},
	"Expires":             {},
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if _, ok := dst[k]; ok {
			if _, shouldSkip := skipHeaders[k]; shouldSkip {
				continue
			}
			if k != "Server" {
				dst.Del(k)
			}
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func pooledIoCopy(dst io.Writer, src io.Reader) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	bufCap := cap(buf)
	io.CopyBuffer(dst, src, buf[0:bufCap:bufCap])
}
