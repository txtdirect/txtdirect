package txtdirect

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

var regexs = []string{
	"^/v2/_catalog$",
	"^/v2/(.*)/tags/(.*)",
	"^/v2/(.*)/manifests/(.*)",
}

func generateDockerv2URI(path string, rec record) (string, error) {
	uri := rec.To
	if rec.Re == "" && rec.To == "" {
		log.Printf("<%s> [txtdirect]: re= and to= fields are required in dockerv2 type.", time.Now().Format(logFormat))
	}
	Dockerv2Regex, err := regexp.Compile(rec.Re)
	if err != nil {
		log.Printf("<%s> [txtdirect]: the given regex doesn't work as expected: %s", time.Now().Format(logFormat), rec.Re)
	}
	pathSubmatches := Dockerv2Regex.FindAllStringSubmatch(path, -1)
	if len(pathSubmatches) < 1 {
		log.Printf("<%s> [txtdirect]: can't extract submatches from %s, with %s", time.Now().Format(logFormat), path, rec.Re)
	}
	pathSlice := pathSubmatches[0][1:]

	for i, v := range pathSlice {
		uri = strings.Replace(uri, fmt.Sprintf("$%d", i+1), v, -1)
	}
	if len(pathSlice) < 1 && rec.Re != "" {
		log.Printf("<%s> [txtdirect]: dockerv2 regex (%s) doesn't work on %s", time.Now().Format(logFormat), rec.Re, path)
	}

	return uri, nil
}
