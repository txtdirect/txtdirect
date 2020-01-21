package record

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"go.txtdirect.org/txtdirect"
	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/variables"
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
		req = test.record.addToContext(req)
		resp := httptest.NewRecorder()
		c := config.Config{
			Redirect: test.redirect,
			Enable:   test.enable,
		}
		fallback(resp, req, test.fallbackType, test.record.Code, c)
		if resp.Code != test.record.Code {
			t.Errorf("Response's status code (%d) doesn't match with expected status code (%d).", resp.Code, test.record.Code)
		}
		if !strings.Contains(resp.Body.String(), test.expected) {
			t.Errorf("Expected response to contain \"%s\".\n\n%s\n\n", test.expected, resp.Body.String())
		}
	}
}

// URLs used are declared in the main zone file in "txtdirect_test.go" file
func Test_fallbackE2E(t *testing.T) {
	tests := []struct {
		url      string
		enable   []string
		code     int
		to       string
		website  string
		root     string
		redirect string
		headers  http.Header
	}{
		{
			url:      "https://127.0.0.1",
			enable:   []string{},
			redirect: "https://isip.test",
			headers:  http.Header{},
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		req.Header = test.headers
		resp := httptest.NewRecorder()
		c := config.Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(variables.Port),
			Enable:   test.enable,
			Redirect: test.redirect,
		}
		err := txtdirect.Redirect(resp, req, c)

		location := resp.Header().Get("Location")

		checkSpecificFallback(t, req, location, test.to, test.website, test.root)

		// Records status code are defined in the txtdirect_test.go file's dns zone
		if resp.Code != test.code {
			checkGlobalFallback(t, req, location, c, resp.Code)
		}

		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	}
}

func checkGlobalFallback(t *testing.T, r *http.Request, location string, config config.Config, code int) {
	if contains(config.Enable, "www") {
		checkLocationHeader(t, location, fmt.Sprintf("https://www.%s", r.URL.Host))
		return
	}
	if config.Redirect != "" {
		checkLocationHeader(t, location, config.Redirect)
		return
	}
	if code != 404 {
		t.Errorf("Expected status code to be 404 but got %d", code)
	}
}

func checkSpecificFallback(t *testing.T, r *http.Request, location, to, website, root string) {
	if to != "" {
		checkLocationHeader(t, location, to)
		return
	}
	if website != "" {
		checkLocationHeader(t, location, website)
		return
	}
	if root != "" {
		checkLocationHeader(t, location, root)
		return
	}
}

func checkLocationHeader(t *testing.T, location, item string) {
	if location != item {
		t.Errorf("Expected %s got %s", item, location)
	}
}
