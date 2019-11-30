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
	headers  http.Header
}

var tests = []test{
	{
		name: "Fallback to \"to=\" when Docker-Client header is not set",
		args: data{
			host: "fallback.dockerv2.dockerv2.example.com",
			path: "/",
		},
		expected: "https://fallback-to.dockerv2.dockerv2.example.com",
	},
	{
		name: "Fallback to \"root=\"",
		args: data{
			host: "fallback.dockerv2.dockerv2.example.com",
			path: "/",
		},
		headers: http.Header{
			"User-Agent": []string{"Docker-Client"},
		},
		expected: "https://fallback-root.dockerv2.dockerv2.example.com",
	},
	{
		name: "Fallback to 404 when \"to=\" is not linked to a Docker image or registry",
		args: data{
			host: "fallback-wrong-to.dockerv2.dockerv2.example.com",
			path: "/v2/test/path",
		},
		headers: http.Header{
			"User-Agent": []string{"Docker-Client"},
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

		req.Header = test.headers

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
	log.Printf("TestCase: \"dockerv2_fallback\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}
