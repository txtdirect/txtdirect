package record

import (
	"net/http/httptest"
	"strings"
	"testing"

	"go.txtdirect.org/txtdirect/config"
)

func Test_fallback(t *testing.T) {
	tests := []struct {
		record       Record
		redirect     string
		fallbackType string
		enable       []string
		url          string
		expected     string
	}{
		{
			record: Record{
				To:   "https://goto.fallback.test",
				Code: 301,
			},
			fallbackType: "to",
			expected:     "https://goto.fallback.test",
		},
		{
			record: Record{
				To:   "",
				Code: 301,
			},
			redirect:     "https://redirect.test",
			fallbackType: "global",
			expected:     "https://redirect.test",
		},
		{
			record: Record{
				To:   "https://goto.fallback.test",
				Code: 404,
			},
			fallbackType: "global",
			expected:     "not found",
		},
		{
			record: Record{
				Website: "https://goto.website.test",
				Code:    302,
			},
			fallbackType: "website",
			expected:     "https://goto.website.test",
		},
		{
			record: Record{
				Root: "https://dockerv2.root.test",
				Code: 302,
			},
			fallbackType: "root",
			expected:     "https://dockerv2.root.test",
		},
		{
			record: Record{
				Code: 302,
			},
			enable:       []string{"www"},
			url:          "https://go.to.www.test",
			fallbackType: "global",
			expected:     "https://www.go.to.www.test",
		},
	}
	for _, test := range tests {
		url := "https://test.test"
		if test.url != "" {
			url = test.url
		}
		req := httptest.NewRequest("GET", url, nil)
		req = test.record.AddToContext(req)
		resp := httptest.NewRecorder()
		c := config.Config{
			Redirect: test.redirect,
			Enable:   test.enable,
		}
		Fallback(resp, req, test.fallbackType, test.record.Code, c)
		if resp.Code != test.record.Code {
			t.Errorf("Response's status code (%d) doesn't match with expected status code (%d).", resp.Code, test.record.Code)
		}
		if !strings.Contains(resp.Body.String(), test.expected) {
			t.Errorf("Expected response to contain \"%s\".\n\n%s\n\n", test.expected, resp.Body.String())
		}
	}
}
