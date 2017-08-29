package txtdirect

import "net/http"

const template = `<!DOCTYPE html>
<head>
<meta name="go-import" content="{{.Host}}{{.Path}} {{.Vcs}} {{.NewUrl}}>
</head>
</html>`

func gometa(w http.ResponseWriter, r record, host, path string) error {
	return nil
}
