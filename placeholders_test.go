package txtdirect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePlaceholders(t *testing.T) {
	tests := []struct {
		url      string
		uri      string
		expected string
	}{
		{
			"example.com{uri}",
			"/?test=test",
			"example.com/?test=test",
		},
		{
			"example.com{uri}/{~test}",
			"/?test=test",
			"example.com/?test=test/test",
		},
		{
			"example.com{uri}/{>Test}",
			"/?test=test",
			"example.com/?test=test/test-header",
		},
		{
			"example.com{uri}/{?test}",
			"/?test=test",
			"example.com/?test=test/test",
		},
		{
			"subdomain.example.com/{label1}",
			"/subdomain",
			"subdomain.example.com/subdomain",
		},
		{
			"subdomain.example.com/{label2}",
			"/example",
			"subdomain.example.com/example",
		},
		{
			"subdomain.example.com/{label3}",
			"/com",
			"subdomain.example.com/com",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", "https://subdomain.example.com"+test.uri, nil)
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

func TestParseLabelLessThanOneFails(t *testing.T) {
	url := "example.com/{label0}"
	uri := "test"
	req := httptest.NewRequest("GET", "https://example.com"+uri, nil)
	_, err := parsePlaceholders(url, req)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseLabelTooHighFails(t *testing.T) {
	url := "example.com/{label9000}"
	uri := "test"
	req := httptest.NewRequest("GET", "https://example.com"+uri, nil)
	_, err := parsePlaceholders(url, req)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
