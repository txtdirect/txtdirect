package main

import (
	"fmt"
	"log"
	"net/http"
)

type data struct {
	host string
	path string
}

type test struct {
	name     string
	args     data
	expected string
	referer  bool
	status   int
}

var tests = []test{
	{
		name: "Fallback to 404",
		args: data{
			host: "fallback.host.host.example.com",
			path: "/",
		},
		status: 404,
	},
}

func main() {
	result := make(map[bool][]test)
	for _, test := range tests {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s", test.args.host, test.args.path), nil)
		if err != nil {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Couldn't create the request: %s", test.name, err.Error())
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Couldn't send the request: %s", test.name, err.Error())
			continue
		}

		// Check response's status code
		if test.status != 0 {
			if resp.StatusCode != test.status {
				result[false] = append(result[false], test)
				log.Printf("[%s]: Expected %d status code, got %d", test.name, test.status, resp.StatusCode)
				continue
			}
			result[true] = append(result[true], test)
			continue
		}

		// Check "Location" header for redirects
		location, err := resp.Location()

		if err == http.ErrNoLocation {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Location header is empty", test.name)
			continue
		}

		if location.String() != test.expected {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Expected %s, got %s", test.name, test.expected, location)
			continue
		}

		// Check "Referer" header if enabled
		if test.referer && resp.Header.Get("Referer") != req.Host {
			result[false] = append(result[false], test)
			log.Printf("Expected %s referer but got \"%s\"", req.Host, resp.Header.Get("Referer"))
			continue
		}

		result[true] = append(result[true], test)
	}
	log.Printf("TestCase: \"host_fallback\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}
