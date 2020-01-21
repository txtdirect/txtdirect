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

package types

import (
	"fmt"
	"go.txtdirect.org/txtdirect/record"
	"net/http/httptest"
	"testing"
)

func Test_zoneFromPath(t *testing.T) {
	tests := []struct {
		url      string
		from     string
		regex    string
		expected string
		err      error
	}{
		{
			"https://example.com/caddy/v1",
			"",
			"",
			"_redirect.v1.caddy.example.com",
			nil,
		},
		{
			"https://example.com/1/2",
			"",
			"",
			"_redirect.2.1.example.com",
			nil,
		},
		{
			"https://example.com/",
			"",
			"",
			"_redirect.example.com",
			nil,
		},
		{
			"https://example.com/caddy/pkg/v1/download",
			"/$1/$4/$2/$3",
			"",
			"_redirect.pkg.download.v1.caddy.example.com",
			nil,
		},
		{
			"https://example.com/caddy/pkg/v1",
			"/$1/$3/$2",
			"",
			"_redirect.pkg.v1.caddy.example.com",
			nil,
		},
		{
			"https://example.com/path-routing",
			"",
			"",
			"_redirect.path-routing.example.com",
			nil,
		},
		{
			"https://example.com/path-routing/test",
			"",
			"",
			"_redirect.test.path-routing.example.com",
			nil,
		},
		{
			"https://example.com/path_routing/test",
			"",
			"",
			"_redirect.test.path_routing.example.com",
			nil,
		},
		{
			"https://example.com/special-chars/#?%!",
			"",
			"",
			"_redirect.special-chars.example.com",
			nil,
		},
		{
			"https://example.com/path2routing/nested/test/",
			"",
			"",
			"_redirect.test.nested.path2routing.example.com",
			nil,
		},
		{
			"https://example.com/123/test",
			"",
			"",
			"_redirect.test.123.example.com",
			nil,
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"",
			"\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)",
			"_redirect.12345-some-path.example.com",
			nil,
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"",
			"(\\d+)",
			"_redirect.12345.example.com",
			nil,
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"",
			"\\?query=([^&]*)",
			"_redirect.string.example.com",
			nil,
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"",
			"\\?query=(?P<a>[^&]+)\\&more=(?P<b>[^&]+)",
			"_redirect.stuff.string.example.com",
			nil,
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff&test=testing",
			"",
			"\\?query=(?P<b>[^&]+)\\&more=(?P<a>[^&]+)\\&test=(?P<c>[^&]+)",
			"_redirect.testing.string.stuff.example.com",
			nil,
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff&test=testing",
			"",
			"\\?query=(?P<a>[^&]+)\\&more=(?P<b2>[^&]+)\\&test=(?P<b1>[^&]+)",
			"_redirect.stuff.testing.string.example.com",
			nil,
		},
		{
			"https://example.com/test",
			"/$2/$1",
			"",
			"",
			fmt.Errorf("length of path doesn't match with length of from= in record"),
		},
	}
	for _, test := range tests {
		rec := record.Record{}
		rec.Re = test.regex
		rec.From = test.from
		req := httptest.NewRequest("GET", test.url, nil)
		zone, _, _, err := zoneFromPath(req, rec)
		if err != nil {
			// Check negative tests
			if err.Error() == test.err.Error() {
				continue
			}
			t.Errorf("Got error: %s", err.Error())
		}
		if zone != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, zone)
		}
	}
}

func Test_zoneFromPathRegexPlaceholders(t *testing.T) {
	tests := []struct {
		url      string
		to       string
		from     string
		regex    string
		expected string
	}{
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"https://example.com/{1}",
			"",
			"\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)",
			"https://example.com/12345-some-path",
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"https://example.com/{1}",
			"",
			"(\\d+)",
			"https://example.com/12345",
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"https://example.com/{1}",
			"",
			"\\?query=([^&]*)",
			"https://example.com/string",
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff",
			"https://example.com/{a}/{b}",
			"",
			"\\?query=(?P<a>[^&]+)\\&more=(?P<b>[^&]+)",
			"https://example.com/string/stuff",
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff&test=testing",
			"https://example.com/{a}/{b}/{c}",
			"",
			"\\?query=(?P<b>[^&]+)\\&more=(?P<a>[^&]+)\\&test=(?P<c>[^&]+)",
			"https://example.com/stuff/string/testing",
		},
		{
			"https://example.com/12345-some-path?query=string&more=stuff&test=testing",
			"https://example.com/{a}/{b2}/{b1}",
			"",
			"\\?query=(?P<a>[^&]+)\\&more=(?P<b2>[^&]+)\\&test=(?P<b1>[^&]+)",
			"https://example.com/string/stuff/testing",
		},
	}
	for _, test := range tests {
		rec := record.Record{
			Re:   test.regex,
			From: test.from,
			To:   test.to,
		}
		req := httptest.NewRequest("GET", test.url, nil)
		_, _, _, err := zoneFromPath(req, rec)
		if err != nil {
			t.Errorf("Unexpected error while parsing path: %s", err.Error())
		}
		to, err := record.ParsePlaceholders(rec.To, req, []string{})
		if err != nil {
			t.Errorf("Unexpected error while parsing placeholders: %s", err.Error())
		}
		if to != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, to)
		}
	}
}
