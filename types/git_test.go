package types

import (
	"net/http/httptest"
	"testing"

	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/record"
)

func TestGit_Proxy(t *testing.T) {
	tests := []struct {
		url     string
		rec     record.Record
		wantErr bool
	}{
		{
			url: "http://git.example.test/info/refs?service=git-upload-pack",
			rec: record.Record{
				To: "https://example.com/example/example.git",
			},
		},
		{
			url: "http://git.example.test/info/refs?service=git-upload-pack",
			rec: record.Record{
				To: "https://example.com/example/example.git",
			},
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		resp := httptest.NewRecorder()
		g := NewGit(resp, req, config.Config{}, test.rec)
		if err := g.Proxy(); err != nil && !test.wantErr {
			t.Errorf("Unexpected error while fetching the Git repository: %s", err.Error())
		}
	}
}
