package txtdirect

import (
	"testing"
)

func Test_zoneFromPath(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		from     string
		expected string
		err      error
	}{
		{
			"example.com",
			"/caddy/v1",
			"",
			"_redirect.v1.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/1/2",
			"",
			"_redirect.2.1.example.com",
			nil,
		},
		{
			"example.com",
			"/",
			"",
			"_redirect.example.com",
			nil,
		},
		{
			"example.com",
			"/caddy/pkg/v1",
			"/$1/$3/$2",
			"_redirect.pkg.v1.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/path-routing",
			"",
			"_redirect.path-routing.example.com",
			nil,
		},
		{
			"example.com",
			"/path-routing/test",
			"",
			"_redirect.test.path-routing.example.com",
			nil,
		},
		{
			"example.com",
			"/path_routing/test",
			"",
			"_redirect.test.path_routing.example.com",
			nil,
		},
		{
			"example.com",
			"/special-chars/#?%!",
			"",
			"_redirect.special-chars.example.com",
			nil,
		},
		{
			"example.com",
			"/path2routing/nested/test/",
			"",
			"_redirect.test.nested.path2routing.example.com",
			nil,
		},
		{
			"example.com",
			"/123/test",
			"",
			"_redirect.test.123.example.com",
			nil,
		},
	}
	for _, test := range tests {
		rec := record{}
		rec.From = test.from
		zone, _, err := zoneFromPath(test.host, test.path, rec)
		if err != nil {
			t.Errorf("Got error: %s", err.Error())
		}
		if zone != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, zone)
		}
	}
}
