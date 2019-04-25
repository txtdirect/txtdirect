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

	"github.com/mholt/caddy/caddyhttp/proxy"
)

type ProxyResponse struct {
	headers    http.Header
	body       []byte
	bodyReader bytes.Buffer
	bodyWriter bytes.Buffer
	status     int
}

func proxyRequest(w http.ResponseWriter, r *http.Request, rec record, c Config, fallbackURL string, code int) error {
	to, _, err := getBaseTarget(rec, r)
	if err != nil {
		return err
	}
	u, err := url.Parse(to)
	if err != nil {
		return err
	}
	reverseProxy := proxy.NewSingleHostReverseProxy(u, "", proxyKeepalive, proxyTimeout, fallbackDelay)

	tmpResponse := ProxyResponse{headers: make(http.Header)}
	reverseProxy.ServeHTTP(&tmpResponse, r, nil)

	// Decompress the body based on "Content-Encoding" header and write to a writer buffer
	if err := tmpResponse.WriteBody(); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the response body: %s", err.Error())
	}

	// Replace the URL hosts with the request's host
	if err := tmpResponse.ReplaceBody(u.Scheme, u.Host, r.Host); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't replace urls inside the response body: %s", err.Error())
	}

	copyHeader(w.Header(), tmpResponse.Header())

	// Write the status from the temporary ResponseWriter to the main ResponseWriter
	w.WriteHeader(tmpResponse.status)

	// Write the final response from the temporary ResponseWriter to the main ResponseWriter
	if _, err := w.Write(tmpResponse.Body()); err != nil {
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

func (p *ProxyResponse) ReplaceBody(scheme, to, host string) error {
	replacedBody := bytes.Replace(p.bodyWriter.Bytes(), []byte(scheme+"://"+to), []byte(scheme+"://"+host), -1)
	p.bodyWriter.Reset()
	if _, err := p.bodyWriter.Write(replacedBody); err != nil {
		return err
	}
	return nil
}

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
