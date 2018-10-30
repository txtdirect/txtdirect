package txtdirect

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_generateDockerv2URI(t *testing.T) {
	tests := []struct {
		path     string
		rec      record
		expected string
	}{
		{
			"/v2/",
			record{
				To:   "docker.io/seetheprogress/txtdirect:latest",
				Code: 302,
			},
			"OK",
		},
		{
			"/v2/testing/tags/docker",
			record{
				To:   "https://gcr.io/v2/my_bucket/$1/tags/$2",
				Code: 302,
			},
			"https://gcr.io/v2/my_bucket/testing/tags/docker",
		},
		{
			"/v2/testing/manifests/docker",
			record{
				To: "https://gcr.io/v2/my_bucket/$1/manifests/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/manifests/docker",
		},
		{
			"/v2/testing/blobs/docker",
			record{
				To: "https://gcr.io/v2/my_bucket/$1/blobs/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/blobs/docker",
		},
		{
			"/v2/_catalog",
			record{
				To: "https://gcr.io/v2/_catalog",
			},
			"https://gcr.io/v2/_catalog",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", fmt.Sprintf("https://example.com%s", test.path), nil)
		resp := httptest.NewRecorder()
		err := redirectDockerv2(resp, req, test.rec)
		if err != nil {
			t.Errorf("Unexpected error happened: %s", err)
		}
		if !strings.Contains(resp.Body.String(), test.expected) {
			t.Errorf("Expected %s, got %s:", test.expected, resp.Body.String())
		}
	}
}
