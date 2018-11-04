package txtdirect

import (
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
			host:     "example.com",
			path:     "/mod/github.com/okkur/reposeed-server/@v/list",
			expected: "",
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		err := gomods(w, test.host, test.path)
		if err != nil {
			t.Errorf("ERROR: %e", err)
		}
	}
}
