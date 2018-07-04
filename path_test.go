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
			"/caddy/pkg/v1/download",
			"/$1/$4/$2/$3",
			"_redirect.v1.pkg.download.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/caddy/pkg/v1",
			"/$1/$3/$2",
			"_redirect.pkg.v1.caddy.example.com",
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
