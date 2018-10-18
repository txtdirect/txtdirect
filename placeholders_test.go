package txtdirect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePlaceholders(t *testing.T) {
	tests := []struct {
		url         string
		placeholder string
		expected    string
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
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", "https://example.com"+test.placeholder, nil)
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

func TestParseSubdomainPlaceholder(t *testing.T) {
	url := "{label1}.example.com"
	placeholder := "kubernetes"
	expected := "kubernetes.example.com"
	req := httptest.NewRequest("GET", "https://"+placeholder+".example.com", nil)
	result, err := parsePlaceholders(url, req)
	if err != nil {
		t.Fatal(err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
