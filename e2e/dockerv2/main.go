package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type data struct {
	host  string
	path  string
	image string
}

type test struct {
	name     string
	args     data
	expected string
	referer  bool
	status   int
	headers  http.Header
	kind     string
	fallback bool
}

var tests = []test{
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
		fallback: true,
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
		status:   404,
		fallback: true,
	},
	{
		name: "Pull an image from the registry",
		args: data{
			image: "pull.dockerv2.dockerv2.example.com/txtdirect",
		},
		kind: "pull",
	},
}

func main() {
	result := make(map[bool][]test)
	for _, test := range tests {
		if test.fallback {
			if err := fallbackTest(test, result); err != nil {
				continue
			}
		}

		switch test.kind {
		case "pull":
			_, err := exec.Command("crane", "pull", "-i", test.args.image, "txtdirect_e2e").CombinedOutput()
			if err != nil {
				result[false] = append(result[false], test)
				log.Printf("Couldn't pull the image from the custom Docker registry: %s", err.Error())
				continue
			}
			_, err = exec.Command("rm", "./txtdirect_e2e").CombinedOutput()
			if err != nil {
				result[false] = append(result[false], test)
				log.Printf("Couldn't remove the pulled image: %s", err.Error())
				continue
			}
		}

		result[true] = append(result[true], test)
	}
	log.Printf("TestCase: \"dockerv2\", Total: %d, Passed: %d, Failed: %d", len(tests), len(result[true]), len(result[false]))
}

// fallbackTest handles the fallback tests and returns an empty non-nil error
// in case of failure.
func fallbackTest(test test, result map[bool][]test) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s", test.args.host, test.args.path), nil)
	if err != nil {
		result[false] = append(result[false], test)
		log.Printf("[%s]: Couldn't create the request: %s", test.name, err.Error())
		return errors.New("")
	}

	req.Header = test.headers

	resp, err := client.Do(req)
	if err != nil {
		result[false] = append(result[false], test)
		log.Printf("[%s]: Couldn't send the request: %s", test.name, err.Error())
		return errors.New("")
	}

	// Check response's status code
	if test.status != 0 {
		if resp.StatusCode != test.status {
			result[false] = append(result[false], test)
			log.Printf("[%s]: Expected %d status code, got %d", test.name, test.status, resp.StatusCode)
			return errors.New("")
		}
		result[true] = append(result[true], test)
		return errors.New("")
	}

	// Check "Location" header for redirects
	location, err := resp.Location()

	if err == http.ErrNoLocation {
		result[false] = append(result[false], test)
		log.Printf("[%s]: Location header is empty", test.name)
		return errors.New("")
	}

	if location.String() != test.expected {
		result[false] = append(result[false], test)
		log.Printf("[%s]: Expected %s, got %s", test.name, test.expected, location)
		return errors.New("")
	}

	// Check "Referer" header if enabled
	if test.referer && resp.Header.Get("Referer") != req.Host {
		result[false] = append(result[false], test)
		log.Printf("Expected %s referer but got \"%s\"", req.Host, resp.Header.Get("Referer"))
		return errors.New("")
	}
	return nil
}
