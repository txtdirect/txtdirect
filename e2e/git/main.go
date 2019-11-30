package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type data struct {
	host string
	path string
	dir  string
}

type test struct {
	name     string
	args     data
	fallback bool
	expected string
}

var tests = []test{
	{
		name: "Clone a repository using a \"git\" record",
		args: data{
			host: "http://package.git.git.example.com",
			path: "",
			dir:  "git",
		},
	},
	{
		name: "Clone a repository using a \"path\" record chained to a \"git\" record",
		args: data{
			host: "http://git.path.example.com",
			path: "/package",
			dir:  "gitpath",
		},
	},
	{
		name: "Fallback to \"website=\" when User-Agent header is not set",
		args: data{
			host: "fallback.git.git.example.com",
			path: "/",
		},
		expected: "https://fallback-website.git.git.example.com",
		fallback: true,
	},
	{
		name: "Fallback to git record's \"website=\" when User-Agent header is not set",
		args: data{
			host: "fallback.git.path.example.com",
			path: "/",
		},
		expected: "https://fallback-website.git.path.example.com",
		fallback: true,
	},
}

func main() {
	result := make(map[bool][]test)
	for _, test := range tests {
		if test.fallback {
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
			continue
		}

		_, err := exec.Command("git", "clone", test.args.host+test.args.path, fmt.Sprintf("/tmp/%s", test.args.dir)).CombinedOutput()
		if err != nil {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Couldn't clone the repository: %s", test.name, err.Error())
			continue
		}

		result[true] = append(result[true], test)
	}
	log.Printf("TestCase: \"git\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}
