/*
Copyright 2017 - The TXTdirect Authors
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
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func dockerv2(w http.ResponseWriter, r *http.Request, rec record) error {
	path := r.URL.Path

	if path == "/v2/" {
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		return err
	}

	dest, err := url.Parse(rec.To)
	if err != nil {
		return fmt.Errorf("Could not parse 'to' URL: %s", err)
	}

	if path == "/v2/_catalog" {
		dest.Path = path
		http.Redirect(w, r, dest.String(), http.StatusMovedPermanently)
		return nil
	}

	exp := regexp.MustCompile("/v2/(.*)/(tags|manifests|blobs)/(.*)")
	if exp.MatchString(path) {
		dst := make([]byte, 0, 1024)
		tpl := "/v2" + strings.TrimSuffix(dest.Path, "/") + "/$1/$2/$3"
		matches := exp.FindStringSubmatchIndex(path)

		dest.Path = string(exp.ExpandString(dst, tpl, path, matches))
		http.Redirect(w, r, dest.String(), http.StatusMovedPermanently)
		return nil
	}

	return fmt.Errorf("Unhandled dockerv2 case")
}
