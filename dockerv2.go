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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var dockerRegexes = map[string]*regexp.Regexp{
	"v2":        regexp.MustCompile("^\\/?v2\\/?$"),
	"container": regexp.MustCompile("v2\\/(([\\w\\d-]+\\/?)+)\\/(tags|manifests|_catalog|blobs)"),
}

type DockerV2 struct {
	Path        string
	Upstream    string
	UpstreamURI *url.URL
}

func (d *DockerV2) Redirect(w http.ResponseWriter, r *http.Request, rec record) error {
	// Handle the "API Version Check" request
	if d.isV2Request(w) {
		return nil
	}

	// Trigger fallback if the request is not a valid Docker v2 API request
	if !d.isValidPath(w, r, rec) {
		log.Printf("[txtdirect]: unrecognized path for dockerv2: %s", d.Path)
		return nil
	}

	if err := d.parseUpstream(); err != nil {
		return err
	}

	// Empty the RequestURI to prevent "http: Request.RequestURI can't be set in client requests" error
	req, _ := http.NewRequest(r.Method, d.UpstreamURI.String(), nil)
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Read the response body and copy headers to the response writer
	respBody, _ := ioutil.ReadAll(resp.Body)
	copyHeader(w.Header(), resp.Header)

	w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	w.Header().Add("Status-Code", strconv.Itoa(http.StatusMovedPermanently))
	w.Write(respBody)

	return nil
}

func (d *DockerV2) parseUpstream() error {
	container := dockerRegexes["container"].FindAllStringSubmatch(d.Path, -1)

	uri, err := url.Parse(d.Upstream)
	if err != nil {
		return err
	}
	d.UpstreamURI = uri

	if d.UpstreamURI.Path == "/" || d.UpstreamURI.Path == "" {
		d.UpstreamURI.Path = d.Path
		return nil
	}

	tag := strings.Split(d.UpstreamURI.Path[1:], ":")
	d.UpstreamURI.Path = strings.Replace(d.Path, container[0][1], tag[0], -1)

	if len(tag) == 2 {
		pathSlice := strings.Split(d.Path, "/")
		pathSlice[len(pathSlice)-1] = tag[1]
		d.UpstreamURI.Path = strings.Join(pathSlice, "/")
	}

	return nil
}

func (d *DockerV2) isValidPath(w http.ResponseWriter, r *http.Request, rec record) bool {
	if !strings.HasPrefix(d.Path, "/v2") {
		if d.Path == "" || d.Path == "/" {
			fallback(w, r, rec.Root, rec.Type, "root", http.StatusPermanentRedirect, Config{})
			return false
		}
		fallback(w, r, rec.Website, rec.Type, "website", http.StatusPermanentRedirect, Config{})
		return false
	}
	return true
}

func (d *DockerV2) isV2Request(w http.ResponseWriter) bool {
	if dockerRegexes["v2"].MatchString(d.Path) {
		w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		w.Write([]byte(http.StatusText(http.StatusOK)))
		return true
	}
	return false
}
