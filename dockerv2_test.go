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
			"/v2/testing/tags/docker",
			record{
				Re: "^/v2/(.*)/tags/(.*)",
				To: "https://gcr.io/v2/my_bucket/$1/tags/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/tags/docker",
		},
		{
			"/v2/testing/manifests/docker",
			record{
				Re: "^/v2/(.*)/manifests/(.*)",
				To: "https://gcr.io/v2/my_bucket/$1/manifests/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/manifests/docker",
		},
		{
			"/v2/testing/blobs/docker",
			record{
				Re: "^/v2/(.*)/blobs/(.*)",
				To: "https://gcr.io/v2/my_bucket/$1/blobs/$2",
			},
			"https://gcr.io/v2/my_bucket/testing/blobs/docker",
		},
		{
			"/v2/_catalog",
			record{
				Re: "^/v2/_catalog$",
				To: "https://gcr.io/v2/_catalog",
			},
			"https://gcr.io/v2/_catalog",
		},
	}
	for _, test := range tests {
		uri, err := generateDockerv2URI(test.path, test.rec)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err.Error())
		}
		if uri != test.expected {
			t.Fatalf("Expected %s, got %s", test.expected, uri)
		}
	}
}
