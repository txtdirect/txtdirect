package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
		name: "Fallback to \"website=\" when User-Agent header is not set",
		args: data{
			host: "fallback.git.git.example.com",
			path: "/",
		},
		expected: "https://fallback-website.git.git.example.com",
	},
	{
		name: "Fallback to git record's \"website=\" when User-Agent header is not set",
		args: data{
			host: "fallback.git.path.example.com",
			path: "/",
		},
		expected: "https://fallback-website.git.path.example.com",
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

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Couldn't read the response body: %s", test.name, err.Error())
			continue
		}

		if !strings.Contains(string(body), test.expected) {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Expected %s to be in %s", test.name, test.expected, string(body))
			continue
		}

		result[true] = append(result[true], test)
	}
	log.Printf("TestCase: \"gometa_fallback\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}
