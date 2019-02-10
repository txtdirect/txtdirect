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
	"html/template"
	"net/http"
	"strings"
)

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="{{.Host}} {{.Vcs}} {{.NewURL}}">
{{if .HasGoSource}}<meta name="go-source" content="{{.Host}} _ {{.NewURL}}/tree/master{/dir} {{.NewURL}}/blob/master{/dir}/{file}#L{line}">{{end}}
</head>
</html>`))

// gometa executes a template on the given ResponseWriter
// that contains go-import meta tag
func gometa(w http.ResponseWriter, r record, host string) error {
	if r.Vcs == "" {
		r.Vcs = "git"
	}

	gosource := strings.Contains(r.To, "github.com")

	RequestsByStatus.WithLabelValues(host, string(http.StatusFound)).Add(1)
	return tmpl.Execute(w, struct {
		Host        string
		Vcs         string
		NewURL      string
		HasGoSource bool
	}{
		host,
		r.Vcs,
		r.To,
		gosource,
	})
}
