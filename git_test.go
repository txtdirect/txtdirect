package txtdirect

import (
	"net/http/httptest"
	"testing"
)

func TestGit_Proxy(t *testing.T) {
	tests := []struct {
		url     string
		rec     record
		wantErr bool
	}{
		{
			url: "http://git.example.test/info/refs?service=git-upload-pack",
			rec: record{
				To: "https://github.com/okkur/torproxy.git",
			},
		},
		{
			url: "http://git.example.test/info/refs?service=git-upload-pack",
			rec: record{
				To: "https://github.com/txtdirect/txtdirect.git",
			},
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		resp := httptest.NewRecorder()
		g := NewGit(resp, req, Config{}, test.rec)
		if err := g.Proxy(); err != nil && !test.wantErr {
			t.Errorf("Unexpected error while fetching the Git repository: %s", err.Error())
		}
	}
}
