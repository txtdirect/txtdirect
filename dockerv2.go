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
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Dockerv2 keeps data for "dockerv2" type requests
type Dockerv2 struct {
	rw  http.ResponseWriter
	req *http.Request
	c   Config
	rec record

	image Image
}

type Image struct {
	Registry string
	Image    string
	Tag      string
}

// NewDockerv2 returns a fresh instance of Dockerv2 struct
func NewDockerv2(w http.ResponseWriter, r *http.Request, rec record, c Config) *Dockerv2 {
	return &Dockerv2{
		rw:  w,
		req: r,
		rec: rec,
		c:   c,
	}
}

var dockerRegexes = map[string]*regexp.Regexp{
	"v2":        regexp.MustCompile("^\\/?v2\\/?$"),
	"container": regexp.MustCompile("/v2/(.*?)\\/\\s*(blobs|manifests)/(.*)"),
}

// Redirect handles the requests for "dockerv2" type
func (d *Dockerv2) Redirect() error {
	path := d.req.URL.Path
	if !strings.HasPrefix(path, "/v2") {
		log.Printf("[txtdirect]: unrecognized path for dockerv2: %s", path)
		if path == "" || path == "/" {
			fallback(d.rw, d.req, "root", http.StatusPermanentRedirect, Config{})
			return nil
		}
		fallback(d.rw, d.req, "website", http.StatusPermanentRedirect, Config{})
		return nil
	}
	if dockerRegexes["v2"].MatchString(path) {
		d.rw.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		_, err := d.rw.Write([]byte(http.StatusText(http.StatusOK)))
		return err
	}
	if path != "/" {
		uri, err := d.ParseReference()
		if err != nil {
			return err
		}
		d.rw.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
		d.rw.Header().Add("Status-Code", strconv.Itoa(http.StatusMovedPermanently))
		http.Redirect(d.rw, d.req, uri, http.StatusMovedPermanently)
		return nil
	}
	d.rw.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	d.rw.Header().Add("Status-Code", strconv.Itoa(http.StatusMovedPermanently))
	http.Redirect(d.rw, d.req, d.rec.To, http.StatusMovedPermanently)
	return nil
}

func (d *Dockerv2) ParseRecordReference() error {
	uri, err := url.Parse(d.rec.To)
	if err != nil {
		return fmt.Errorf("Couldn't parse the record endpoint: %s", err.Error())
	}

	if uri.Path == "" {
		uri.Path = "/"
	}

	path := strings.Split(uri.Path, ":")
	if len(path) == 2 {
		d.image.Tag = path[1]
	}

	d.image.Registry = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)

	d.image.Image = path[0][1:]

	return nil
}

func (d *Dockerv2) ParseReference() (string, error) {
	path := d.req.URL.Path
	if err := d.ParseRecordReference(); err != nil {
		return "", fmt.Errorf("Couldn't parse the image reference: %s", err.Error())
	}

	matches := dockerRegexes["container"].FindAllStringSubmatch(path, -1)[0]

	// Replace image tag on manifests requests
	if matches[2] == "manifests" {
		if d.image.Tag != "" {
			path = strings.Replace(path, matches[3], d.image.Tag, -1)
		}
	}

	// Replace image name and namepsace
	if d.image.Image != "" {
		path = strings.Replace(path, matches[1], d.image.Image, -1)
	}

	return fmt.Sprintf("%s%s", d.image.Registry, path), nil
}
