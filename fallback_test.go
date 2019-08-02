package txtdirect

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func Test_fallback(t *testing.T) {
	tests := []struct {
		record       record
		redirect     string
		fallbackType string
		enable       []string
		url          string
		expected     string
	}{
		{
			record: record{
				To:   "https://goto.fallback.test",
				Code: 301,
			},
			fallbackType: "to",
			expected:     "https://goto.fallback.test",
		},
		{
			record: record{
				To:   "",
				Code: 301,
			},
			redirect:     "https://redirect.test",
			fallbackType: "global",
			expected:     "https://redirect.test",
		},
		{
			record: record{
				To:   "https://goto.fallback.test",
				Code: 404,
			},
			fallbackType: "global",
			expected:     "not found",
		},
		{
			record: record{
				Website: "https://goto.website.test",
				Code:    302,
			},
			fallbackType: "website",
			expected:     "https://goto.website.test",
		},
		{
			record: record{
				Root: "https://dockerv2.root.test",
				Code: 302,
			},
			fallbackType: "root",
			expected:     "https://dockerv2.root.test",
		},
		{
			record: record{
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
		req = addRecordToContext(req, test.record)
		resp := httptest.NewRecorder()
		c := Config{
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
			url:     "https://fallbackpath.test/to",
			enable:  []string{"path"},
			code:    302,
			to:      "https://to.works.fine.test",
			headers: http.Header{},
		},
		{
			url:     "https://fallbackpath.test",
			enable:  []string{"www"},
			headers: http.Header{},
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		req.Header = test.headers
		resp := httptest.NewRecorder()
		c := Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(port),
			Enable:   test.enable,
			Redirect: test.redirect,
		}
		err := Redirect(resp, req, c)

		checkSpecificFallback(t, resp, req, test.to, test.website, test.root)

		// Records status code are defined in the txtdirect_test.go file's dns zone
		if resp.Code != test.code {
			checkGlobalFallback(t, resp, req, c)
		}

		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	}
}

func checkGlobalFallback(t *testing.T, resp *httptest.ResponseRecorder, r *http.Request, config Config) {
	if contains(config.Enable, "www") {
		checkLocationHeader(t, resp, fmt.Sprintf("https://www.%s", r.URL.Host))
		return
	}
	if config.Redirect != "" {
		checkLocationHeader(t, resp, config.Redirect)
		return
	}
	if resp.Code != 404 {
		t.Errorf("Expected status code to be 404 but got %d", resp.Code)
	}
}

func checkSpecificFallback(t *testing.T, resp *httptest.ResponseRecorder, r *http.Request, to, website, root string) {
	if to != "" {
		checkLocationHeader(t, resp, to)
		return
	}
	if website != "" {
		checkLocationHeader(t, resp, website)
		return
	}
	if root != "" {
		checkLocationHeader(t, resp, root)
		return
	}
}

func checkLocationHeader(t *testing.T, resp *httptest.ResponseRecorder, item string) {
	if resp.Header().Get("Location") != item {
		t.Errorf("Expected %s got %s", item, resp.Header().Get("Location"))
	}
}
