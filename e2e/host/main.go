package host

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
		name: "Redirect to a host record's to= field",
		args: data{
			host: "to.host.host.example.com",
			path: "/",
		},
		expected: "http://to-redirect.host.host.example.com",
	},
}

func Run() error {
	for _, test := range tests {
		resp, err := http.Get(fmt.Sprintf("http://%s%s", test.args.host, test.args.path))
		if err != nil {
			return fmt.Errorf("Couldn't send the request: %s", err.Error())
		}

		location, err := resp.Location()

		fmt.Println(resp.Request.Header)
		if err == http.ErrNoLocation {
			return fmt.Errorf("[%s]: Location header is empty", test.name)
		}

		if location.String() != test.expected {
			return fmt.Errorf("[%s]: Expected %s, got %s", test.name, test.expected, location)
		}
	}
	return nil
}
