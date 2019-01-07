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
	"html/template"
	"net/http"
	"strings"
)

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="{{.Host}}{{.Path}} {{.Vcs}} {{.NewURL}}">
<meta name="go-source" content="{{.Host}}{{.Path}} _ {{.NewURL}}/tree/v2{/dir} {{.NewURL}}/blob/v2{/dir}/{file}#L{line}">
</head>
</html>`))

// gometa executes a template on the given ResponseWriter
// that contains go-import meta tag
func gometa(w http.ResponseWriter, r record, host, path string) error {
	if r.Vcs == "" {
		r.Vcs = "git"
	}
	if path == "/" {
		path = ""
	}
	bl := "/internal"
	if strings.Contains(path, bl) {
		return fmt.Errorf("path containing 'internal' is disallowed")
	}
	RequestsByStatus.WithLabelValues(host, string(http.StatusFound)).Add(1)
	return tmpl.Execute(w, struct {
		Host   string
		Path   string
		Vcs    string
		NewURL string
	}{
		host,
		path,
		r.Vcs,
		r.To,
	})
}
