package txtdirect

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var PathRegex = regexp.MustCompile("\\/([A-Za-z0-9-._~!$'()*+,;=:@]+)")
var FromRegex = regexp.MustCompile("\\/\\$(\\d+)")

func zoneFromPath(host string, path string, rec record) (string, int, error) {
	pathSubmatchs := PathRegex.FindAllStringSubmatch(path, -1)
	if rec.Re != "" {
		CustomRegex, err := regexp.Compile(rec.Re)
		if err != nil {
			log.Printf("<%s> [txtdirect]: the given regex doesn't work as expected: %s", time.Now().String(), rec.Re)
		}
		pathSubmatchs = CustomRegex.FindAllStringSubmatch(path, -1)
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

func reverse(input []string) {
	last := len(input) - 1
	for i := 0; i < len(input)/2; i++ {
		input[i], input[last-i] = input[last-i], input[i]
	}
}
