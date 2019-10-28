package path

import (
	"fmt"
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
		expected: "https://noroot-redicte.path.path.example.com",
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
}

func Run() error {
	for _, test := range tests {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s", test.args.host, test.args.path), nil)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Couldn't send the request: %s", err.Error())
		}

		location, err := resp.Location()

		if err == http.ErrNoLocation {
			return fmt.Errorf("[%s]: Location header is empty", test.name)
		}

		if location.String() != test.expected {
			return fmt.Errorf("[%s]: Expected %s, got %s", test.name, test.expected, location)
		}
	}
	return nil
}
