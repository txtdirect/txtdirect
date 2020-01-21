/*
Copyright 2019 - The TXTDirect Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package record

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePlaceholders(t *testing.T) {
	tests := []struct {
		url       string
		requested string
		pathSlice []string
		expected  string
	}{
		{
			"example.com{uri}",
			"https://example.com/about/test/file.html?query=string&another=value#even-a-fragment",
			[]string{},
			"example.com/about/test/file.html?query=string&another=value#even-a-fragment",
		},
		{
			"example.com{uri_escaped}",
			"https://example.com/about/test/file.html?query=string&another=value#even-a-fragment",
			[]string{},
			"example.com%2Fabout%2Ftest%2Ffile.html%3Fquery%3Dstring%26another%3Dvalue%23even-a-fragment",
		},
		{
			"example.com/{~test}",
			"https://example.com/?test=test",
			[]string{},
			"example.com/test",
		},
		{
			"example.com/{>Test}",
			"https://example.com/?test=test",
			[]string{},
			"example.com/test-header",
		},
		{
			"example.com/{?test}",
			"https://example.com/?test=test",
			[]string{},
			"example.com/test",
		},
		{
			"example.com{dir}",
			"https://example.com/directory/test",
			[]string{},
			"example.com/directory/",
		},
		{
			"example.com/directory/{file}",
			"https://example.com/directory/file.pdf",
			[]string{},
			"example.com/directory/file.pdf",
		},
		{
			"example.com/{host}",
			"https://project.example.com:8080",
			[]string{},
			"example.com/project.example.com:8080",
		},
		{
			"example.com/{hostonly}",
			"https://project.example.com",
			[]string{},
			"example.com/project.example.com",
		},
		{
			"example.com/{method}",
			"https://example.com",
			[]string{},
			"example.com/GET",
		},
		{
			"example.com{path}",
			"https://example.com/this-is-a-test-path/api/v1/",
			[]string{},
			"example.com/this-is-a-test-path/api/v1/",
		},
		{
			"example.com{path_escaped}",
			"https://example.com/path_escaped-test/api/v1",
			[]string{},
			"example.com%2Fpath_escaped-test%2Fapi%2Fv1",
		},
		{
			"example.com:{port}",
			"https://example.com:8080",
			[]string{},
			"example.com:8080",
		},
		{
			"example.com/test/{query}",
			"https://example.com/test/?querykey=queryvalue&anotherquerykey=anothervalue",
			[]string{},
			"example.com/test/querykey=queryvalue&anotherquerykey=anothervalue",
		},
		{
			"example.com/test/{query_escaped}",
			"https://example.com/test/?querykey=queryvalue&anotherquerykey=anothervalue",
			[]string{},
			"example.com/test/querykey%3Dqueryvalue%26anotherquerykey%3Danothervalue",
		},
		{
			"example.com/{user}",
			"https://example.com/user1",
			[]string{},
			"example.com/user1",
		},
		{
			"about.example.com/{label1}",
			"https://about.example.com",
			[]string{},
			"about.example.com/about",
		},
		{
			"about.example.com/{label2}",
			"https://about.example.com",
			[]string{},
			"about.example.com/example",
		},
		{
			"about.example.com/{label3}",
			"https://about.example.com/com",
			[]string{},
			"about.example.com/com",
		},
		{
			"about.example.com/{$1}/{$3}/{$2}",
			"https://about.example.com/this/is/test",
			[]string{"this", "is", "test"},
			"about.example.com/this/test/is",
		},
		{
			"about.example.com/{$1}/{$3}/{$2}",
			"https://about.example.com/123/test/t3st",
			[]string{"123", "test", "t3st"},
			"about.example.com/123/t3st/test",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.requested, nil)
		req.AddCookie(&http.Cookie{Name: "test", Value: "test"})
		req.Header.Add("Test", "test-header")
		req.SetBasicAuth("user1", "password")
		result, err := ParsePlaceholders(test.url, req, test.pathSlice)
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
		pathSlice []string
		requested string
	}{
		{
			"example.com/{label0}",
			[]string{},
			"https://example.com/test",
		},
		{
			"example.com/{label9000}",
			[]string{},
			"https://example.com/test",
		},
		{
			"example.com/{labelxyz}",
			[]string{},
			"https://example.com/test",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.requested, nil)
		_, err := ParsePlaceholders(test.url, req, test.pathSlice)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}
}
