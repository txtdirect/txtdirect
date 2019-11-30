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
		name: "Redirect a gometa record to a Go package",
		args: data{
			host: "gometa.gometa.example.com",
			path: "?go-get=1",
		},
		expected: "https://gometa-redirect.gometa.gometa.example.com/golang/package",
	},
	{
		name: "Redirect to a Go package using a gometa record chained to a path record",
		args: data{
			host: "path.example.com",
			path: "/gometa?go-get=1",
		},
		expected: "https://gometa-redirect.gometa.path.example.com/golang/package",
	},
	{
		name: "Redirect to a Go package using a path record with multiple chained gometa records",
		args: data{
			host: "path.example.com",
			path: "/gometa/multiple?go-get=1",
		},
		expected: "https://gometa-redirect.gometa.path.example.com/second/golang/package",
	},
	{
		name: "Fallback to path record's \"to=\" when gometa's \"to=\" is empty",
		args: data{
			host: "fallback.gometa.path.example.com",
			path: "/fallback-empty-to",
		},
		expected: "https://fallback-to.gometa.path.example.com",
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
	log.Printf("TestCase: \"gometa\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}
