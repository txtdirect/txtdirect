package txtdirect

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var PathRegex = regexp.MustCompile("\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)")
var FromRegex = regexp.MustCompile("\\/\\$(\\d+)")

func zoneFromPath(host string, path string, rec record) (string, int, error) {
	pathSubmatchs := PathRegex.FindAllStringSubmatch(path, -1)
	pathSlice := []string{}
	for _, v := range pathSubmatchs {
		pathSlice = append(pathSlice, v[1])
	}
	from := len(pathSlice)
	if rec.From != "" {
		fromSubmatch := FromRegex.FindAllStringSubmatch(rec.From, -1)
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
			return "", 0, fmt.Errorf("length of path doesn't match with length of from= in record")
		}
		generatedPath := []string{}

		sort.Sort(sort.Reverse(sort.IntSlice(keys)))

		for _, k := range keys {
			generatedPath = append(generatedPath, fromSlice[k])
		}

		url := append(generatedPath, host)
		url = append([]string{basezone}, url...)
		return strings.Join(url, "."), from, nil
	}
	reverse(pathSlice)
	url := append(pathSlice, host)
	url = append([]string{basezone}, url...)
	return strings.Join(url, "."), from, nil
}

func getFinalRecord(zone string, from int) (record, error) {
	var txts []string
	var err error

	txts, err = net.LookupTXT(zone)

	// if nothing found, jump into wildcards
	for i := 1; i <= from && len(txts) == 0; i++ {
		zoneSlice := strings.Split(zone, ".")
		zoneSlice[i] = "_"
		zone = strings.Join(zoneSlice, ".")
		txts, err = net.LookupTXT(zone)
	}
	if err != nil || len(txts) == 0 {
		return record{}, fmt.Errorf("could not get TXT record: %s", err)
	}

	rec := record{}
	if err = rec.Parse(txts[0]); err != nil {
		return rec, fmt.Errorf("could not parse record: %s", err)
	}

	return rec, nil
}

func reverse(input []string) {
	last := len(input) - 1
	for i := 0; i < len(input)/2; i++ {
		input[i], input[last-i] = input[last-i], input[i]
	}
}
