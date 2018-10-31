package txtdirect

import (
	"fmt"
	"net/http"
)

func gomods(w http.ResponseWriter, host, path string, rec record) error {
	if rec.Vcs == "" {
		rec.Vcs = "mod"
	}
	if path == "/" {
		path = ""
	}

	url := fmt.Sprintf("%s@%s", rec.To, rec.ModVersion)

	return tmpl.Execute(w, struct {
		Host   string
		Path   string
		Vcs    string
		NewURL string
	}{
		host,
		path,
		rec.Vcs,
		url,
	})
}
