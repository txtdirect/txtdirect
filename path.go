package txtdirect

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func zoneFromPath(host string, path string, rec record) (string, int, error) {
	match, err := regexp.Compile("([a-zA-Z0-9]+)")
	if err != nil {
		return "", 0, err
	}
	pathSlice := match.FindAllString(path, -1)
	from := len(pathSlice)
	if rec.From != "" {
		match, err := regexp.Compile("\\/\\$(\\d+)")
		if err != nil {
			return "", 0, err
		}
		fromSlice := match.FindAllStringSubmatch(rec.From, -1)
		generatedPath := []string{}
		for _, v := range fromSlice {
			index, _ := strconv.Atoi(v[1])
			generatedPath = append(generatedPath, pathSlice[index-1])
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
