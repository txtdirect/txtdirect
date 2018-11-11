package txtdirect

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func Test_gomods(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		expected string
	}{
		{
			path:     "/github.com/okkur/reposeed-server/@v/list",
			expected: "https://github.com/okkur/reposeed-server/@v/list",
		},
		{
			path:     "/github.com/okkur/reposeed-server/@v/v1.0.0.info",
			expected: "https://github.com/okkur/reposeed-server/@v/v1.0.0.info",
		},
		{
			path:     "/github.com/okkur/reposeed-server/@v/v1.0.0.mod",
			expected: "https://github.com/okkur/reposeed-server/@v/v1.0.0.mod",
		},
		{
			path:     "/github.com/okkur/reposeed-server/@v/v1.0.0.zip",
			expected: "https://github.com/okkur/reposeed-server/@v/v1.0.0.zip",
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com%s", test.path), nil)
		c := Config{}
		c.ModProxy.Cache.Enable = true
		c.ModProxy.Cache.Type = "local"
		c.ModProxy.Cache.Path = "/home/erbesharat/.test"
		err := gomods(w, r, test.path, c)
		if err != nil {
			t.Errorf("ERROR: %e", err)
		}
		if r.URL.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, r.URL.String())
		}
	}
}
