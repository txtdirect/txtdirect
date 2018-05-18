package txtdirect

import (
	"testing"
)

func Test_zoneFromPath(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		expected string
		err      error
	}{
		{
			"example.com",
			"/caddy/v1",
			"_redirect.v1.caddy.example.com",
			nil,
		},
		{
			"example.com",
			"/1/2",
			"_redirect.2.1.example.com",
			nil,
		},
		{
			"example.com",
			"/",
			"_redirect.example.com",
			nil,
		},
	}
	for _, test := range tests {
		zone, _, err := zoneFromPath(test.host, test.path)
		if err != nil {
			t.Errorf("Got error: %s", err.Error())
		}
		if zone != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, zone)
		}
	}
}
