/*
Copyright 2017 - The TXTDirect Authors
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

package types

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/plugins/prometheus"
	"go.txtdirect.org/txtdirect/record"
)

// Gometa keeps data for "gometa" type requests
type Gometa struct {
	rw  http.ResponseWriter
	req *http.Request
	c   config.Config
	rec record.Record
}

// NewGometa returns a fresh instance of Gometa struct
func NewGometa(w http.ResponseWriter, r *http.Request, rec record.Record, c config.Config) *Gometa {
	return &Gometa{
		rw:  w,
		req: r,
		rec: rec,
		c:   c,
	}
}

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="{{.Host}}{{.Path}} {{.Vcs}} {{.NewURL}}">
{{if .HasGoSource}}<meta name="go-source" content="{{.Host}}{{.Path}} _ {{.NewURL}}/tree/master{/dir} {{.NewURL}}/blob/master{/dir}/{file}#L{line}">{{end}}
</head>
</html>`))

// Serve executes a template on the given ResponseWriter
// that contains go-import meta tag
func (g *Gometa) Serve() error {
	if g.rec.Vcs == "" {
		g.rec.Vcs = "git"
	}
	if g.req.URL.Path == "/" {
		g.req.URL.Path = ""
	}

	gosource := strings.Contains(g.rec.To, "github.com")

	prometheus.RequestsByStatus.WithLabelValues(g.req.Host, strconv.Itoa(http.StatusFound)).Add(1)
	return tmpl.Execute(g.rw, struct {
		Host        string
		Path        string
		Vcs         string
		NewURL      string
		HasGoSource bool
	}{
		g.req.Host,
		g.req.URL.Path,
		g.rec.Vcs,
		g.rec.To,
		gosource,
	})
}

// ValidQuery checks the request query to make sure the requests are
// coming from the Go tool.
func (g *Gometa) ValidQuery() bool {
	if g.req.URL.Query().Get("go-get") != "1" {
		record.Fallback(g.rw, g.req, "website", http.StatusFound, g.c)
		return false
	}
	return true
}
