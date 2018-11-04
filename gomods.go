package txtdirect

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type ModProxy struct {
	Enable bool
	Path   string
	Cache  string
}

func gomods(w http.ResponseWriter, host, path string, c Config) error {
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

func modExists(name string, c Config) bool {
	path := fmt.Sprintf("%s/%s", c.ModProxy.Cache, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func createModDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directory: %s", path)
	}
	return nil
}
