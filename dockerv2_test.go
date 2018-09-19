package txtdirect

import (
	"testing"
)

func Test_generateDockerv2URI(t *testing.T) {
	tests := []struct {
		path     string
		rec      record
		expected string
	}{
		{
			"",
			record{
				To:   "docker.io/seetheprogress/txtdirect:latest",
				Code: 302,
			},
			"docker.io/seetheprogress/txtdirect:latest",
		},
	}
	for _, test := range tests {
		uri, _ := generateDockerv2URI(test.path, test.rec)
		if uri != test.expected {
			t.Fatalf("Expected %s, got %s", test.expected, uri)
		}
	}
}
