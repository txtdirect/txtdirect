package txtdirect

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_gomods(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		expected string
	}{
		{
			path: "/github.com/okkur/reposeed-server/@v/list",
		},
		{
			path: "/github.com/okkur/reposeed-server/@v/v0.1.0.info",
		},
		{
			path: "/github.com/okkur/reposeed-server/@v/v0.1.0.mod",
		},
		{
			path: "/github.com/okkur/reposeed-server/@v/v0.1.0.zip",
		},
	}
	for _, test := range tests {
		if err := os.MkdirAll("/tmp/gomods", os.ModePerm); err != nil {
			t.Fatal("Couldn't create storage directory (/tmp/gomods)")
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com%s", test.path), nil)
		c := Config{
			Gomods: Gomods{
				Enable:   true,
				Workers:  2,
				GoBinary: os.Getenv("GOROOT") + "/bin/go",
				Cache: struct {
					Type string
					Path string
				}{
					Type: "local",
					Path: "/tmp/gomods",
				},
			},
		}
		err := gomods(w, r, test.path, c)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	}
}
