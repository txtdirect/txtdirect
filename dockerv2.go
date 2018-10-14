package txtdirect

import (
	"fmt"
	"regexp"
	"strings"
)

var dockerRegexs = map[string]string{
	"_catalog":  "^/v2/_catalog$",
	"tags":      "^/v2/(.*)/tags/(.*)",
	"manifests": "^/v2/(.*)/manifests/(.*)",
	"blobs":     "^/v2/(.*)/blobs/(.*)",
}

var DockerRegex = regexp.MustCompile("^\\/v2\\/(.*\\/(tags|manifests|blobs)\\/.*|_catalog$)")

func generateDockerv2URI(path string, rec record) (string, int) {
	if path != "/" {
		regexType := DockerRegex.FindAllStringSubmatch(path, -1)[0]
		pathRegex, err := regexp.Compile(dockerRegexs[regexType[len(regexType)-1]])
		if err != nil {
			panic(err)
		}
		pathSubmatches := pathRegex.FindAllStringSubmatch(path, -1)
		pathSlice := pathSubmatches[0][1:]

		uri := rec.To
		for i, v := range pathSlice {
			uri = strings.Replace(uri, fmt.Sprintf("$%d", i+1), v, -1)
		}
		return uri, rec.Code
	}
	return rec.To, rec.Code
}
