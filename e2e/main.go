package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var directories []string
	if err := listDirectories(&directories); err != nil {
		log.Printf("[txtdirect_e2e]: Couldn't list the test directories: %s", err.Error())
	}
	fmt.Println(directories)
}

func listDirectories(directories *[]string) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Couldn't find the directory %q: %s", path, err.Error())
		}
		if info.IsDir() && path != "." {
			*directories = append(*directories, path)
			return filepath.SkipDir
		}
		return nil
	})
}
