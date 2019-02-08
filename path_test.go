package txtdirect

import (
	"fmt"
	"testing"
)

func Test_zoneFromPath(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		from     string
		regex    string
		expected string
		err      error
	}{
		{
			"example.com",
			"/caddy/v1",
			"",
			"",
			"_redirect.v1.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/1/2",
			"",
			"",
			"_redirect.2.1.example.com",
			nil,
		},
		{
			"example.com",
			"/",
			"",
			"",
			"_redirect.example.com",
			nil,
		},
		{
			"example.com",
			"/caddy/pkg/v1/download",
			"/$1/$4/$2/$3",
			"",
			"_redirect.pkg.download.v1.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/caddy/pkg/v1",
			"/$1/$3/$2",
			"",
			"_redirect.pkg.v1.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/path-routing",
			"",
			"",
			"_redirect.path-routing.example.com",
			nil,
		},
		{
			"example.com",
			"/path-routing/test",
			"",
			"",
			"_redirect.test.path-routing.example.com",
			nil,
		},
		{
			"example.com",
			"/path_routing/test",
			"",
			"",
			"_redirect.test.path_routing.example.com",
			nil,
		},
		{
			"example.com",
			"/special-chars/#?%!",
			"",
			"",
			"_redirect.special-chars.example.com",
			nil,
		},
		{
			"example.com",
			"/path2routing/nested/test/",
			"",
			"",
			"_redirect.test.nested.path2routing.example.com",
			nil,
		},
		{
			"example.com",
			"/123/test",
			"",
			"",
			"_redirect.test.123.example.com",
			nil,
		},
		{
			"example.com",
			"/12345-some-path?query=string&more=stuff",
			"",
			"\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)",
			"_redirect.12345-some-path.example.com",
			nil,
		},
		{
			"example.com",
			"/12345-some-path?query=string&more=stuff",
			"",
			"(\\d+)",
			"_redirect.12345.example.com",
			nil,
		},
		{
			"example.com",
			"/12345-some-path?query=string&more=stuff",
			"",
			"\\?query=([^&]*)",
			"_redirect.string.example.com",
			nil,
		},
		{
			"example.com",
			"/12345-some-path?query=string&more=stuff",
			"",
			"\\?query=(?P<a>[^&]+)\\&more=(?P<b>[^&]+)",
			"_redirect.stuff.string.example.com",
			nil,
		},
		{
			"example.com",
			"/12345-some-path?query=string&more=stuff&test=testing",
			"",
			"\\?query=(?P<b>[^&]+)\\&more=(?P<a>[^&]+)\\&test=(?P<c>[^&]+)",
			"_redirect.testing.string.stuff.example.com",
			nil,
		},
		{
			"example.com",
			"/12345-some-path?query=string&more=stuff&test=testing",
			"",
			"\\?query=(?P<a>[^&]+)\\&more=(?P<b2>[^&]+)\\&test=(?P<b1>[^&]+)",
			"_redirect.stuff.testing.string.example.com",
			nil,
		},
		{
			"example.com",
			"/test",
			"/$2/$1",
			"",
			"",
			fmt.Errorf("length of path doesn't match with length of from= in record"),
		},
	}
	for _, test := range tests {
		rec := record{}
		rec.Re = test.regex
		rec.From = test.from
		zone, _, _, err := zoneFromPath(test.host, test.path, rec)
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
