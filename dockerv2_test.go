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
				To:   "https://gcr.io/seetheprogress/txtdirect:latest",
				Code: 302,
			},
			"OK",
		},
		{
			"/v2/random/container/tags/latest",
			record{
				To:   "https://gcr.io/",
				Code: 302,
			},
			"https://gcr.io/v2/random/container/tags/latest",
		},
		{
			"/v2/random/container/tags/v2.0.0",
			record{
				To:   "https://gcr.io/",
				Code: 302,
			},
			"https://gcr.io/v2/random/container/tags/v2.0.0",
		},
		{
			"/v2/testing/container/tags/v3.0.0",
			record{
				To:   "https://gcr.io/testing/container:v2.0.0",
				Code: 302,
			},
			"https://gcr.io/v2/testing/container/tags/v2.0.0",
		},
		{
			"/v2/testing/container/tags/v2.0.0",
			record{
				To:   "https://gcr.io/testing/container",
				Code: 302,
			},
			"https://gcr.io/v2/testing/container/tags/v2.0.0",
		},
		{
			"/v2/random/container/tags/latest",
			record{
				To:   "https://gcr.io/testing/container:v2.0.0",
				Code: 302,
			},
			"https://gcr.io/v2/testing/container/tags/v2.0.0",
		},

		{
			"/v2/random/container/tags/v2.0.0",
			record{
				To:   "https://gcr.io/testing/container",
				Code: 302,
			},
			"https://gcr.io/v2/testing/container/tags/v2.0.0",
		},
		{
			"/v2/random/container/_catalog",
			record{
				To:   "https://gcr.io/",
				Code: 302,
			},
			"https://gcr.io/v2/random/container/_catalog",
		},
		{
			"/random/path",
			record{
				To:      "https://gcr.io/",
				Code:    302,
				Website: "https://fallback.test",
			},
			"https://fallback.test",
		},
		{
			"",
			record{
				To:   "https://gcr.io/",
				Code: 302,
				Root: "https://fallback.test",
			},
			"https://fallback.test",
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
