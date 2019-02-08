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

var PathRegex = regexp.MustCompile("\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)")
var FromRegex = regexp.MustCompile("\\/\\$(\\d+)")
var GroupRegex = regexp.MustCompile("P<[a-zA-Z]+[a-zA-Z0-9]*>")
var GroupOrderRegex = regexp.MustCompile("P<([a-zA-Z]+[a-zA-Z0-9]*)>")

// zoneFromPath generates a DNS zone with the given host and path
// It will use custom regex to parse the path if it's provided in
// the given record.
func zoneFromPath(host string, path string, rec record) (string, int, []string, error) {
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
			reverse(url)
			from := len(pathSlice)
			url = append(url, host)
			url = append([]string{basezone}, url...)
			return strings.Join(url, "."), from, pathSlice, nil
		}
	}
	pathSlice := []string{}
	for _, v := range pathSubmatchs {
		pathSlice = append(pathSlice, v[1])
	}
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

		url := append(generatedPath, host)
		url = append([]string{basezone}, url...)
		return strings.Join(url, "."), from, pathSlice, nil
	}
	ps := pathSlice
	reverse(pathSlice)
	url := append(pathSlice, host)
	url = append([]string{basezone}, url...)
	return strings.Join(url, "."), from, ps, nil
}

// getFinalRecord finds the final TXT record for the given zone.
// It will try wildcards if the first zone return error
func getFinalRecord(zone string, from int, ctx context.Context, c Config, r *http.Request, pathSlice []string) (record, error) {
	txts, err := query(zone, ctx, c)
	if err != nil {
		// if nothing found, jump into wildcards
		for i := 1; i <= from && len(txts) == 0; i++ {
			zoneSlice := strings.Split(zone, ".")
			zoneSlice[i] = "_"
			zone = strings.Join(zoneSlice, ".")
			txts, err = query(zone, ctx, c)
		}
	}
	if err != nil || len(txts) == 0 {
		return record{}, fmt.Errorf("could not get TXT record: %s", err)
	}

	txts[0], err = parsePlaceholders(txts[0], r, pathSlice)
	rec := record{}
	if err = rec.Parse(txts[0], r, c); err != nil {
		return rec, fmt.Errorf("could not parse record: %s", err)
	}

	if rec.Type == "path" {
		return rec, fmt.Errorf("chaining path is not currently supported")
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
