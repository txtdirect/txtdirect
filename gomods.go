package txtdirect

import (
	"log"
	"net/http"
	"strings"
)

type ModProxy struct {
	Enable bool
	Path   string
	Cache  string
}

func gomods(w http.ResponseWriter, host, path string) error {
	pathSlice := strings.Split(path, "/")[2:] // [2:] ignores proxy's base url and the empty slice item
	var moduleName string
	var fileName string
	for k, v := range pathSlice {
		if v == "@v" {
			fileName = pathSlice[k+1]
			break
		}
		moduleName = strings.Join([]string{moduleName, v}, "/")
	}
	log.Println(fileName)
	return nil
}
