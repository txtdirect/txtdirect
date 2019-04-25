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

var dockerRegexes = map[string]*regexp.Regexp{
	"v2":        regexp.MustCompile("^\\/?v2\\/?$"),
	"container": regexp.MustCompile("v2\\/(([\\w\\d-]+\\/?)+)\\/(tags|manifests|_catalog|blobs)"),
}

func redirectDockerv2(w http.ResponseWriter, r *http.Request, rec record) error {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/v2") {
		log.Printf("[txtdirect]: unrecognized path for dockerv2: %s", path)
		if path == "" || path == "/" {
			fallback(w, r, rec.Root, rec.Type, "root", http.StatusPermanentRedirect, Config{})
			return nil
		}
		fallback(w, r, rec.Website, rec.Type, "website", http.StatusPermanentRedirect, Config{})
		return nil
	}
	if dockerRegexes["v2"].MatchString(path) {
		w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		return err
	}
	if path != "/" {
		uri, err := createDockerv2URI(rec.To, path)
		if err != nil {
			return err
		}
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
		w.Header().Add("Status-Code", strconv.Itoa(http.StatusMovedPermanently))
		http.Redirect(w, r, uri, http.StatusMovedPermanently)
		return nil
	}
	w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	w.Header().Add("Status-Code", strconv.Itoa(http.StatusMovedPermanently))
	http.Redirect(w, r, rec.To, http.StatusMovedPermanently)
	return nil
}

func createDockerv2URI(to string, path string) (string, error) {
	uri, err := url.Parse(to)
	if err != nil {
		return "", err
	}

	if uri.Path == "/" || uri.Path == "" {
		uri.Path = path
		return uri.String(), nil
	}

	// Replace container's path in docker's request with what's inside rec.To
	containerPath := dockerRegexes["container"].FindAllStringSubmatch(path, -1)[0][1] // [0][1]: The second item in first group is always container path
	containerAndVersion := strings.Split(uri.Path, ":")                               // First item in slice is container and second item is version
	uri.Path = strings.Replace(path, containerPath, containerAndVersion[0][1:], -1)

	// Replace the version number in docker's request with what's inside rec.To
	if len(containerAndVersion) == 2 {
		pathSlice := strings.Split(uri.Path, "/")
		pathSlice[len(pathSlice)-1] = containerAndVersion[1]
		uri.Path = strings.Join(pathSlice, "/")
	}

	return uri.String(), nil
}
