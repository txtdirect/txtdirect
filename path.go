/*
Copyright 2019 - The TXTDirect Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package txtdirect

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Path contains the data that are needed to redirect path requests
type Path struct {
	rw   http.ResponseWriter
	req  *http.Request
	c    Config
	path string
	rec  record
}

var PathRegex = regexp.MustCompile("\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)")
var FromRegex = regexp.MustCompile("\\/\\$(\\d+)")
var GroupRegex = regexp.MustCompile("P<[a-zA-Z]+[a-zA-Z0-9]*>")
var GroupOrderRegex = regexp.MustCompile("P<([a-zA-Z]+[a-zA-Z0-9]*)>")

// NewPath returns an instance of Path struct using the given data
func NewPath(w http.ResponseWriter, r *http.Request, path string, rec record, c Config) *Path {
	return &Path{
		rw:   w,
		req:  r,
		path: path,
		rec:  rec,
		c:    c,
	}
}

// Redirect finds and returns the final record
func (p *Path) Redirect() *record {
	zone, from, pathSlice, err := zoneFromPath(p.req, p.rec)
	rec, err := getFinalRecord(zone, from, p.c, p.rw, p.req, pathSlice)
	*p.req = *rec.addToContext(p.req)
	if err != nil {
		log.Print("Fallback is triggered because an error has occurred: ", err)
		fallback(p.rw, p.req, "to", p.rec.Code, p.c)
		return nil
	}

	if rec.Type == "path" {
		p.rec = rec
		return p.Redirect()
	}

	return &rec
}

// RedirectRoot redirects the request to record's root= field
// if the path is empty or "/". If the root= field is empty too
// fallback will be triggered.
func (p *Path) RedirectRoot() error {
	if p.rec.Root == "" {
		fallback(p.rw, p.req, "to", p.rec.Code, p.c)
		return nil
	}
	log.Printf("[txtdirect]: %s > %s", p.req.Host+p.req.URL.Path, p.rec.Root)
	if p.rec.Code == http.StatusMovedPermanently {
		p.rw.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	}
	p.rw.Header().Add("Status-Code", strconv.Itoa(p.rec.Code))
	http.Redirect(p.rw, p.req, p.rec.Root, p.rec.Code)
	if p.c.Prometheus.Enable {
		RequestsByStatus.WithLabelValues(p.req.Host, strconv.Itoa(p.rec.Code)).Add(1)
	}
	return nil
}

// zoneFromPath generates a DNS zone with the given request's path and host
// It will use custom regex to parse the path if it's provided in
// the given record.
func zoneFromPath(r *http.Request, rec record) (string, int, []string, error) {
	path := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)

	if strings.ContainsAny(path, ".") {
		path = strings.Replace(path, ".", "-", -1)
	}
	pathSubmatchs := PathRegex.FindAllStringSubmatch(path, -1)
	if rec.Re != "" {
		CustomRegex, err := regexp.Compile(rec.Re)
		if err != nil {
			log.Printf("<%s> [txtdirect]: the given regex doesn't work as expected: %s", time.Now().String(), rec.Re)
		}
		pathSubmatchs = CustomRegex.FindAllStringSubmatch(path, -1)
		if GroupRegex.MatchString(rec.Re) {
			pathSlice := []string{}
			unordered := make(map[string]string)
			for _, item := range pathSubmatchs[0] {
				pathSlice = append(pathSlice, item)
			}
			order := GroupOrderRegex.FindAllStringSubmatch(rec.Re, -1)
			for i, group := range order {
				unordered[group[1]] = pathSlice[i+1]
			}
			url := sortMap(unordered)
			*r = *r.WithContext(context.WithValue(r.Context(), "regexMatches", unordered))
			reverse(url)
			from := len(pathSlice)
			url = append(url, r.Host)
			url = append([]string{basezone}, url...)
			return strings.Join(url, "."), from, pathSlice, nil
		}
	}
	pathSlice := []string{}
	for _, v := range pathSubmatchs {
		pathSlice = append(pathSlice, v[1])
	}
	*r = *r.WithContext(context.WithValue(r.Context(), "regexMatches", pathSlice))
	if len(pathSlice) < 1 && rec.Re != "" {
		log.Printf("<%s> [txtdirect]: custom regex doesn't work on %s", time.Now().String(), path)
	}
	from := len(pathSlice)
	if rec.From != "" {
		fromSubmatch := FromRegex.FindAllStringSubmatch(rec.From, -1)
		if len(fromSubmatch) != len(pathSlice) {
			return "", 0, []string{}, fmt.Errorf("length of path doesn't match with length of from= in record")
		}
		fromSlice := make(map[int]string)
		for k, v := range fromSubmatch {
			index, _ := strconv.Atoi(v[1])
			fromSlice[index] = pathSlice[k]
		}

		keys := []int{}
		for k := range fromSlice {
			keys = append(keys, k)
		}
		if len(keys) != len(pathSlice) {
			return "", 0, []string{}, fmt.Errorf("length of path doesn't match with length of from= in record")
		}
		generatedPath := []string{}

		sort.Sort(sort.Reverse(sort.IntSlice(keys)))

		for _, k := range keys {
			generatedPath = append(generatedPath, fromSlice[k])
		}

		url := append(generatedPath, r.Host)
		url = append([]string{basezone}, url...)
		return strings.Join(url, "."), from, pathSlice, nil
	}
	ps := pathSlice
	reverse(pathSlice)
	url := append(pathSlice, r.Host)
	url = append([]string{basezone}, url...)
	return strings.Join(url, "."), from, ps, nil
}

// getFinalRecord finds the final TXT record for the given zone.
// It will try wildcards if the first zone return error
func getFinalRecord(zone string, from int, c Config, w http.ResponseWriter, r *http.Request, pathSlice []string) (record, error) {
	txts, err := query(zone, r.Context(), c)
	if err != nil {
		// if nothing found, jump into wildcards
		for i := 1; i <= from && len(txts) == 0; i++ {
			zoneSlice := strings.Split(zone, ".")
			zoneSlice[i] = "_"
			zone = strings.Join(zoneSlice, ".")
			txts, err = query(zone, r.Context(), c)
		}
	}
	if err != nil || len(txts) == 0 {
		return record{}, fmt.Errorf("could not get TXT record: %s", err)
	}

	txts[0], err = parsePlaceholders(txts[0], r, pathSlice)
	rec := record{}
	if err = rec.Parse(txts[0], w, r, c); err != nil {
		return rec, fmt.Errorf("could not parse record: %s", err)
	}

	if rec.Type == "path" {
		records := r.Context().Value("records").([]record)
		parent := records[len(records)-1]

		// Use the parent's custom regex if available
		if rec.Re == "" && parent.Re != "" {
			rec.Re = parent.Re
		}

		return rec, nil
	}

	return rec, nil
}

// reverse reverses the order of the array
func reverse(input []string) {
	last := len(input) - 1
	for i := 0; i < len(input)/2; i++ {
		input[i], input[last-i] = input[last-i], input[i]
	}
}

func sortMap(m map[string]string) []string {
	var keys []string
	var result []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, m[k])
	}
	return result
}
