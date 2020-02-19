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

// RegexRecord holds the TXT record and re= field of a predefined regex record
type RegexRecord struct {
	Position   int
	TXT        string
	Regex      string
	Submatches []string
}

// PathRegex is the default regex to parse request's path
// It can be replaced using the re= field in the records
var PathRegex = regexp.MustCompile("\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)")

// FromRegex parses the from= field
var FromRegex = regexp.MustCompile("\\/\\$(\\d+)")

// GroupRegex parses the re= field to find the regex groups
var GroupRegex = regexp.MustCompile("P<[a-zA-Z]+[a-zA-Z0-9]*>")

// GroupOrderRegex finds the order of regex groups inside re=
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
	log.Printf("[txtdirect]: %s > %s", UpstreamZone(p.req)+p.req.URL.Path, p.rec.Root)
	if p.rec.Code == http.StatusMovedPermanently {
		p.rw.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	}
	p.rw.Header().Add("Status-Code", strconv.Itoa(p.rec.Code))
	http.Redirect(p.rw, p.req, p.rec.Root, p.rec.Code)
	if p.c.Prometheus.Enable {
		RequestsByStatus.WithLabelValues(UpstreamZone(p.req), strconv.Itoa(p.rec.Code)).Add(1)
	}
	return nil
}

// SpecificRecord finds the most specific match using the custom regexes from subzones
// It goes through all the custom regexes specified in each subzone and uses the
// most specific match to return the final record.
func (p *Path) SpecificRecord() (*record, error) {
	// Iterate subzones and parse the records
	regexes, err := p.fetchRegexes()
	if err != nil {
		return nil, err
	}

	// Find the most specific regex
	record, err := p.specificMatch(regexes)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (p *Path) specificMatch(regexes []RegexRecord) (*record, error) {
	var specificZone RegexRecord
	// Run each regex on the path and list them in a map
	for _, zone := range regexes {
		// Compile the regex and find the path submatches
		regex, err := regexp.Compile(zone.Regex)
		if err != nil {
			return nil, fmt.Errorf("Couldn't compile the regex: %s", err.Error())
		}
		matches := regex.FindAllStringSubmatch(p.path, -1)

		if len(matches) == 0 {
			continue
		}

		zone.Submatches = matches[0]

		// Use the next regex if it has more matches
		if len(zone.Submatches) > len(specificZone.Submatches) {
			specificZone = zone
		}
	}

	// Add the most specific match's path slice to the request context to use in placeholders
	*p.req = *p.req.WithContext(context.WithValue(p.req.Context(), "regexMatches", specificZone.Submatches))

	// Parse the specific regex record
	var rec record
	var err error
	if rec, err = ParseRecord(specificZone.TXT, p.rw, p.req, p.c); err != nil {
		return nil, fmt.Errorf("Could not parse record: %s", err)
	}

	return &rec, nil
}

func (p *Path) fetchRegexes() ([]RegexRecord, error) {
	regexes := []RegexRecord{}
	for i, loop := 1, true; loop != false; i++ {
		// Send a DNS query to each predefined regex record
		txts, err := query(fmt.Sprintf("%d.%s", i, UpstreamZone(p.req)), p.req.Context(), p.c)
		if err != nil && len(regexes) >= 1 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Couldn't fetch the subzones for predefined regex: %s", err.Error())
		}

		// All predefined regex records should contain the re= field (even non-path records)
		if !strings.Contains(txts[0], "re=") {
			return nil, fmt.Errorf("Couldn't find the re= field in records: %s", err.Error())
		}

		// Extract the re= field from record and add it to the regex slice
		regexes = append(regexes, RegexRecord{
			Position: i,
			TXT:      txts[0],
			Regex:    strings.TrimPrefix(strings.Split(txts[0][strings.Index(txts[0], "re="):], ";")[0], "re="),
		})
	}

	sort.Slice(regexes, func(i, j int) bool {
		return regexes[i].Position < regexes[j].Position
	})

	return regexes, nil
}

// zoneFromPath generates a DNS zone with the given request's path and host
// It will use custom regex to parse the path if it's provided in
// the given record.
func zoneFromPath(r *http.Request, rec record) (string, int, []string, error) {
	path := r.URL.Path

	// Check if request is from a Git client
	if strings.HasPrefix(r.Header.Get("User-Agent"), "git") {
		path = path[:strings.Index(path, "/info/refs")]
	}

	path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)

	pathSubmatchs := PathRegex.FindAllStringSubmatch(path, -1)
	if rec.Re != "" {
		// Compile the record regex and find path submatches
		CustomRegex, err := regexp.Compile(rec.Re)
		if err != nil {
			log.Printf("<%s> [txtdirect]: the given regex doesn't work as expected: %s", time.Now().String(), rec.Re)
		}
		pathSubmatchs = CustomRegex.FindAllStringSubmatch(path, -1)

		// Only generate the zone if the custom regex contains a group
		if GroupRegex.MatchString(rec.Re) {
			//
			pathSlice := []string{}
			unordered := make(map[string]string)
			for _, item := range pathSubmatchs[0] {
				pathSlice = append(pathSlice, item)
			}

			// Order the path slice using groups order in custom regex
			order := GroupOrderRegex.FindAllStringSubmatch(rec.Re, -1)
			for i, group := range order {
				unordered[group[1]] = pathSlice[i+1]
			}

			url := sortMap(unordered)
			*r = *r.WithContext(context.WithValue(r.Context(), "regexMatches", unordered))
			url = normalize(url)
			reverse(url)
			from := len(pathSlice)
			url = append(url, UpstreamZone(r))
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
		return "", 0, []string{}, fmt.Errorf("custom regex doesn't work on %s", path)
	}
	pathSlice = normalize(pathSlice)
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

		url := append(generatedPath, UpstreamZone(r))
		url = append([]string{basezone}, url...)
		return strings.Join(url, "."), from, pathSlice, nil
	}
	ps := pathSlice
	reverse(pathSlice)
	url := append(pathSlice, UpstreamZone(r))
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
	var rec record
	if rec, err = ParseRecord(txts[0], w, r, c); err != nil {
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

// Normalize the path to follow RFC1034 rules
func normalize(input []string) []string {
	var result []string
	for _, value := range input {
		if strings.ContainsAny(value, ".") {
			value = strings.Replace(value, ".", "-", -1)
		}
		result = append(result, value)
	}
	return result
}
