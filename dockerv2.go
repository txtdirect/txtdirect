package txtdirect

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var dockerRegexes = map[string]*regexp.Regexp{
	"v2":        regexp.MustCompile("^\\/?v2\\/?$"),
	"container": regexp.MustCompile("v2\\/(([\\w\\d-]+\\/?)+)\\/(tags|manifests|_catalog|blobs)"),
}

func redirectDockerv2(w http.ResponseWriter, r *http.Request, rec record) error {
	path := r.URL.Path
	if dockerRegexes["v2"].MatchString(path) {
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		return err
	}
	if path != "/" {
		uri, err := createDockerv2URI(rec.To, path)
		if err != nil {
			return err
		}
		http.Redirect(w, r, uri, http.StatusMovedPermanently)
		return nil
	}
	http.Redirect(w, r, rec.To, http.StatusMovedPermanently)
	return nil
}

func createDockerv2URI(to string, path string) (string, error) {
	uri, err := url.Parse(to)
	if err != nil {
		return "", err
	}

	if uri.Path == "/" {
		uri.Path = path
		return uri.String(), nil
	}

	// Replace container's path in docker's request with what's inside rec.To
	containerPath := dockerRegexes["container"].FindAllStringSubmatch(path, -1)[0][1] // [0][1]: The second item in first group is always container path
	containerAndVersion := strings.Split(uri.Path, ":")                               // First item in slice is container and second item is version
	uri.Path = strings.Replace(path, containerPath, containerAndVersion[0][1:], -1)

	// Replace the version number in docker's request with what's inside rec.To
	if len(containerAndVersion) == 2 {
		pathSlice := strings.Split(uri.Path, "/")
		pathSlice[len(pathSlice)-1] = containerAndVersion[1]
		uri.Path = strings.Join(pathSlice, "/")
	}

	return uri.String(), nil
}
