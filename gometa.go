package txtdirect

import (
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<head>
<meta name="go-import" content="{{.Host}}{{.Path}} {{.Vcs}} {{.NewURL}}">
</head>
</html>`))

func gometa(w http.ResponseWriter, r record, host, path string) error {
	if path == "/" {
		path = ""
	}

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
