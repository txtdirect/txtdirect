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

package txtdirect

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Gometa struct {
	rw  http.ResponseWriter
	req *http.Request
	c   Config
	rec record
}

func NewGometa(w http.ResponseWriter, r *http.Request, rec record, c Config) *Gometa {
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
<meta name="go-import" content="{{.Host}}{{.Path}} {{.Vcs}} {{.PackageURI}}">
<meta name="go-import" content="{{.Host}}{{.Path}} mod https://{{.PackageURI}}">
{{if .HasGoSource}}<meta name="go-source" content="{{.Host}}{{.Path}} _ {{.PackageURI}}/tree/master{/dir} {{.PackageURI}}/blob/master{/dir}/{file}#L{line}">{{end}}
</head>
</html>`))

// gometa executes a template on the given ResponseWriter
// that contains go-import meta tag
func (g *Gometa) Serve() error {
	if g.rec.Vcs == "" {
		g.rec.Vcs = "git"
	}
	if g.req.URL.Path == "/" {
		g.req.URL.Path = ""
	}

	gosource := strings.Contains(g.rec.To, "github.com")

	url, err := url.Parse(g.rec.To)
	if err != nil {
		return fmt.Errorf("Unable to parse package's URL: %s", err.Error())
	}

	RequestsByStatus.WithLabelValues(g.req.Host, strconv.Itoa(http.StatusFound)).Add(1)
	return tmpl.Execute(g.rw, struct {
		Host        string
		Path        string
		Vcs         string
		PackageURI  string
		ImportPath  string
		HasGoSource bool
	}{
		g.req.Host,
		g.req.URL.Path,
		g.rec.Vcs,
		url.String(),
		url.Host + url.Path,
		gosource,
	})
}

func (g *Gometa) ValidQuery() bool {
	if g.req.URL.Query().Get("go-get") != "1" {
		fallback(g.rw, g.req, "website", http.StatusFound, g.c)
		return false
	}
	return true
}
