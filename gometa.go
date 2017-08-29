package txtdirect

import (
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<head>
<meta name="go-import" content="{{.Host}}{{.Path}} {{.Vcs}} {{.NewUrl}}">
</head>
</html>`))

func gometa(w http.ResponseWriter, r record, host, path string) error {
	return tmpl.Execute(w, struct {
		Host string
		Path string
		Vcs  string
		URL  string
	}{
		host,
		path,
		r.Vcs,
		r.To,
	})
}
