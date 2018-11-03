package txtdirect

import (
	"net/http"
)

type ModProxy struct {
	Path  string
	Cache string
}

func gomods(w http.ResponseWriter, host, path string, rec record) error {
	return nil
}
