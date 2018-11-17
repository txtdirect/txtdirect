package txtdirect

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var dockerRegexes = map[string]*regexp.Regexp{
	"v2":        regexp.MustCompile("^\\/?v2\\/?$"),
	"_catalog":  regexp.MustCompile("^/v2/_catalog$"),
	"tags":      regexp.MustCompile("^/v2/(.*)/tags/(.*)"),
	"manifests": regexp.MustCompile("^/v2/(.*)/manifests/(.*)"),
	"blobs":     regexp.MustCompile("^/v2/(.*)/blobs/(.*)"),
}

var DockerRegex = regexp.MustCompile("^\\/v2\\/(.*\\/(tags|manifests|blobs)\\/.*|_catalog$)")

func redirectDockerv2(w http.ResponseWriter, r *http.Request, rec record) error {
	path := r.URL.Path
	if dockerRegexes["v2"].MatchString(path) {
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		return err
	}
	if path != "/" {
		regexType := DockerRegex.FindAllStringSubmatch(path, -1)[0]
		if regexType[len(regexType)-1] == "" {
			regexType = regexType[:len(regexType)-1]
		}
		pathSubmatches := dockerRegexes[regexType[len(regexType)-1]].FindAllStringSubmatch(path, -1)
		pathSlice := pathSubmatches[0][1:]

		uri := rec.To + path
		for i, v := range pathSlice {
			uri = strings.Replace(uri, fmt.Sprintf("$%d", i+1), v, -1)
		}
		http.Redirect(w, r, uri, http.StatusMovedPermanently)
		return nil
	}
	http.Redirect(w, r, rec.To, http.StatusMovedPermanently)
	return nil
}
