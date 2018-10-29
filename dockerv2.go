package txtdirect

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var dockerRegexs = map[string]string{
	"v2":        "^\\/?v2\\/?$",
	"_catalog":  "^/v2/_catalog$",
	"tags":      "^/v2/(.*)/tags/(.*)",
	"manifests": "^/v2/(.*)/manifests/(.*)",
	"blobs":     "^/v2/(.*)/blobs/(.*)",
}

var DockerRegex = regexp.MustCompile("^\\/v2\\/(.*\\/(tags|manifests|blobs)\\/.*|_catalog$)")

func redirectDockerv2(w http.ResponseWriter, r *http.Request, rec record) error {
	path := r.URL.Path
	v2Regex, err := regexp.Compile(dockerRegexs["v2"])
	if err != nil {
		panic(err)
	}
	if v2Regex.MatchString(path) {
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		return err
	}
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
		http.Redirect(w, r, uri, http.StatusMovedPermanently)
		return nil
	}
	http.Redirect(w, r, rec.To, http.StatusMovedPermanently)
	return nil
}
