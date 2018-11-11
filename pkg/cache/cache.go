package cache

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

var files = []string{
	"/@v/list",
	"/@v/{version}.info",
	"/@v/{version}.mod",
	"/@v/{version}.zip",
}

func Local(url, localPath, version string) error {
	path := strings.Join([]string{localPath, version}, "/")
	if _, err := os.Stat(path + "/" + version); !os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create directory: %s", path)
		}
		return nil
	}
	fetch(url, localPath, version)
	return nil
}

func fetch(module, localPath, version string) error {
	for _, file := range files {
		var content []byte
		u := strings.Replace(file, "{module}", module, -1)
		u = strings.Replace(u, "{version}", version, -1)
		panic(strings.Join([]string{localPath, u}, "/"))
		file, err := os.Create(strings.Join([]string{localPath, u}, "/"))
		if err != nil {
			return err
		}
		defer file.Close()
		resp, err := http.Get(u)
		if err != nil {
			return err
		}
		resp.Body.Read(content)
		file.Write(content)
	}
	return nil
}
