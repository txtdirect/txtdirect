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
}

var tests = []test{
	{
		name: "Redirect a path record without specified v=",
		args: data{
			host: "noversion.path.path.example.com",
			path: "/",
		},
		expected: "https://noversion-redirect.path.path.example.com",
	},
	{
		name: "Redirect a path record without specified root=",
		args: data{
			host: "noroot.path.path.example.com",
			path: "/",
		},
		expected: "https://noroot-redirect.path.path.example.com",
	},
	{
		name: "Redirect a path record using predefined regex records",
		args: data{
			host: "predefined-regex.path.path.example.com",
			path: "/test1",
		},
		expected: "https://predefined-redirect.host.path.example.com/first/test1",
	},
	{
		name: "Redirect a path record using predefined regex records",
		args: data{
			host: "predefined-regex.path.path.example.com",
			path: "/test1/test2",
		},
		expected: "https://predefined-redirect.host.path.example.com/second/test1/test2",
	},
	{
		name: "Fallback to \"to=\" when wildcard not found",
		args: data{
			host: "path.path.example.com",
			path: "/not/found",
		},
		expected: "https://fallback-to.path.path.example.com",
	},
	{
		name: "Fallback to \"to=\" when \"re=\" and \"from=\" both exist",
		args: data{
			host: "fallback-refrom.path.path.example.com",
			path: "/",
		},
		expected: "https://fallback-to.path.path.example.com",
	},
	{
		name: "Fallback to \"to=\" when \"from=\" length doesn't match request's path length",
		args: data{
			host: "fallback-lenfrom.path.path.example.com",
			path: "/",
		},
		expected: "https://fallback-to.path.path.example.com",
	},
	{
		name: "Redirect using the use= field and upstream record",
		args: data{
			host: "sourcerecord.path.path.example.com",
			path: "/testing",
		},
		expected: "https://upstream.path.path.example.com",
	},
	{
		name: "Redirect to upstream record using multiple use= fields",
		args: data{
			host: "srcrecord.path.path.example.com",
			path: "/testing",
		},
		expected: "https://upstream.path.path.example.com",
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
		resp, err := client.Do(req)
		if err != nil {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Couldn't send the request: %s", test.name, err.Error())
			continue
		}

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

		result[true] = append(result[true], test)
	}
	log.Printf("TestCase: \"path\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}
