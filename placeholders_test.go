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
			"https://example.com/about/test/file.html?query=string&another=value#even-a-fragment",
			"example.com/about/test/file.html?query=string&another=value#even-a-fragment",
		},
		{
			"example.com{uri_escaped}",
			"https://example.com/about/test/file.html?query=string&another=value#even-a-fragment",
			"example.com%2Fabout%2Ftest%2Ffile.html%3Fquery%3Dstring%26another%3Dvalue%23even-a-fragment",
		},
		{
			"example.com/{~test}",
			"https://example.com/?test=test",
			"example.com/test",
		},
		{
			"example.com/{>Test}",
			"https://example.com/?test=test",
			"example.com/test-header",
		},
		{
			"example.com/{?test}",
			"https://example.com/?test=test",
			"example.com/test",
		},
		{
			"example.com{dir}",
			"https://example.com/directory/test",
			"example.com/directory/",
		},
		{
			"example.com/directory/{file}",
			"https://example.com/directory/file.pdf",
			"example.com/directory/file.pdf",
		},
		{
			"example.com/{host}",
			"https://project.example.com:8080",
			"example.com/project.example.com:8080",
		},
		{
			"example.com/{hostonly}",
			"https://project.example.com",
			"example.com/project.example.com",
		},
		{
			"example.com/{method}",
			"https://example.com",
			"example.com/GET",
		},
		{
			"example.com{path}",
			"https://example.com/this-is-a-test-path/api/v1/",
			"example.com/this-is-a-test-path/api/v1/",
		},
		{
			"example.com{path_escaped}",
			"https://example.com/path_escaped-test/api/v1",
			"example.com%2Fpath_escaped-test%2Fapi%2Fv1",
		},
		{
			"example.com:{port}",
			"https://example.com:8080",
			"example.com:8080",
		},
		{
			"example.com/test/{query}",
			"https://example.com/test/?querykey=queryvalue&anotherquerykey=anothervalue",
			"example.com/test/querykey=queryvalue&anotherquerykey=anothervalue",
		},
		{
			"example.com/test/{query_escaped}",
			"https://example.com/test/?querykey=queryvalue&anotherquerykey=anothervalue",
			"example.com/test/querykey%3Dqueryvalue%26anotherquerykey%3Danothervalue",
		},
		{
			"example.com/{user}",
			"https://example.com/user1",
			"example.com/user1",
		},
		{
			"about.example.com/{label1}",
			"https://about.example.com",
			"about.example.com/about",
		},
		{
			"about.example.com/{label2}",
			"https://about.example.com",
			"about.example.com/example",
		},
		{
			"about.example.com/{label3}",
			"https://about.example.com/com",
			"about.example.com/com",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.requested, nil)
		req.AddCookie(&http.Cookie{Name: "test", Value: "test"})
		req.Header.Add("Test", "test-header")
		req.SetBasicAuth("user1", "password")
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
