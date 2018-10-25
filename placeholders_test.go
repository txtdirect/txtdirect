package txtdirect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePlaceholders(t *testing.T) {
	tests := []struct {
		url       string
		requested string
		expected  string
	}{
		{
			"example.com{uri}",
			"https://example.com/?test=test",
			"example.com/?test=test",
		},
		{
			"example.com{uri}/{~test}",
			"https://example.com/?test=test",
			"example.com/?test=test/test",
		},
		{
			"example.com{uri}/{>Test}",
			"https://example.com/?test=test",
			"example.com/?test=test/test-header",
		},
		{
			"example.com{uri}/{?test}",
			"https://example.com/?test=test",
			"example.com/?test=test/test",
		},
		{
			"example.com{dir}",
			"https://example.com/directory/test",
			"example.com/directory/",
		},
		{
			"subdomain.example.com/{label1}",
			"https://subdomain.example.com/subdomain",
			"subdomain.example.com/subdomain",
		},
		{
			"subdomain.example.com/{label2}",
			"https://subdomain.example.com/example",
			"subdomain.example.com/example",
		},
		{
			"subdomain.example.com/{label3}",
			"https://subdomain.example.com/com",
			"subdomain.example.com/com",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.requested, nil)
		req.AddCookie(&http.Cookie{Name: "test", Value: "test"})
		req.Header.Add("Test", "test-header")
		result, err := parsePlaceholders(test.url, req)
		if err != nil {
			t.Fatal(err)
		}
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestParsePlaceholdersFails(t *testing.T) {
	tests := []struct {
		url       string
		requested string
	}{
		{
			"example.com/{label0}",
			"https://example.com/test",
		},
		{
			"example.com/{label9000}",
			"https://example.com/test",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.requested, nil)
		_, err := parsePlaceholders(test.url, req)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}
}
