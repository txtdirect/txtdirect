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
		{
			"/v2/testing/tags/docker",
			record{
				To:   "https://gcr.io/v2/my_bucket/$1/tags/$2",
				Code: 302,
			},
			"https://gcr.io/v2/my_bucket/testing/tags/docker",
		},
		{
			"/v2/testing/manifests/docker",
			record{
				To: "https://gcr.io/v2/my_bucket/$1/manifests/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/manifests/docker",
		},
		{
			"/v2/testing/blobs/docker",
			record{
				To: "https://gcr.io/v2/my_bucket/$1/blobs/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/blobs/docker",
		},
		{
			"/v2/_catalog",
			record{
				To: "https://gcr.io/v2/_catalog",
			},
			"https://gcr.io/v2/_catalog",
		},
	}
	for _, test := range tests {
		uri, _ := generateDockerv2URI(test.path, test.rec)
		if uri != test.expected {
			t.Fatalf("Expected %s, got %s", test.expected, uri)
		}
	}
}
