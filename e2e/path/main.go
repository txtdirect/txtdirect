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
	fallback bool
	status   int
	comment  string
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
		expected: "https://predefined-redirect.host.path.example.com/second/test1",
	},
	{
		name: "Redirect a path record using predefined regex records",
		args: data{
			host: "predefined-regex.path.path.example.com",
			path: "/test1/test2",
		},
		expected: "https://predefined-redirect.host.path.example.com/first/test1/test2",
	},
	{
		name: "Fallback to \"to=\" when wildcard not found",
		args: data{
			host: "path.path.example.com",
			path: "/not/found",
		},
		fallback: true,
		expected: "https://fallback-to.path.path.example.com",
	},
	{
		name: "Fallback to \"to=\" when \"re=\" and \"from=\" both exist",
		args: data{
			host: "fallback-refrom.path.path.example.com",
			path: "/",
		},
		fallback: true,
		expected: "https://fallback-to.path.path.example.com",
	},
	{
		name: "Fallback to \"to=\" when \"from=\" length doesn't match request's path length",
		args: data{
			host: "fallback-lenfrom.path.path.example.com",
			path: "/",
		},
		fallback: true,
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
		name: "Redirect to upstream record using multiple use= fields and skipping non-working upstream",
		args: data{
			host: "src-multiple-use.path.path.example.com",
			path: "/testing",
		},
		expected: "https://upstream.path.path.example.com",
	},
	{
		name: "Fallback when the upstream record doesn't respond",
		args: data{
			host: "fallbackupstream.path.path.example.com",
			path: "/testing",
		},
		fallback: true,
		status:   404,
	},
	{
		name: "Prioritise the parent record's to= field to chained records in fallback",
		args: data{
			host: "pathchain.path.path.example.com",
			path: "/",
		},
		expected: "https://fallback-to.path.path.example.com",
	},
	{
		name: "Fallback when wrong path is given to chained path records",
		args: data{
			host: "pathchain.path.path.example.com",
			path: "/wrong",
		},
		expected: "https://fallback-unknown-path.path.path.example.com",
	},
	{
		name: "Use regex placeholders in a host record chained to a path record",
		args: data{
			host: "chaining.path.path.example.com",
			path: "/host/",
		},
		expected: "https://redirect-host.host.path.example.com",
	},
	{
		name: "Use regex placeholders in fallback triggered by chained record",
		args: data{
			host: "chaining-regex.path.path.example.com",
			path: "/fallback-placeholder/",
		},
		expected: "https://fallback.path.path.example.com/fallback-placeholder",
	},
	{
		name: "Fallback to root= when path is empty for custom regex",
		args: data{
			host: "numbered-regex.host.path.example.com",
			path: "",
		},
		expected: "https://fallback.host.path.example.com",
		comment:  "for: apt.k8s.io",
	},
	{
		name: "Redirect to host record's to= field using numbered regex",
		args: data{
			host: "numbered-regex.host.path.example.com",
			path: "/testing",
		},
		expected: "https://package.host.path.example.com/apt/testing",
		comment:  "for: apt.k8s.io",
	},
	{
		name: "Redirect to root= when path is / for custom regex",
		args: data{
			host: "custom-numbered.host.path.example.com",
			path: "/",
		},
		expected: "https://index.host.path.example.com",
		comment:  "for: changelog.k8s.io",
	},
	{
		name: "Redirect to host record's to= for custom regex using placeholders",
		args: data{
			host: "custom-numbered.host.path.example.com",
			path: "/testing",
		},
		expected: "https://redirect.host.path.example.com/testing",
		comment:  "for: changelog.k8s.io",
	},
	{
		name: "Fallback to root= when path is empty for predefined regexes",
		args: data{
			host: "predefined-numbered.host.path.example.com",
			path: "/",
		},
		expected: "https://index.host.path.example.com",
		comment:  "for: ci-test.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains digits for predefined regex",
		args: data{
			host: "predefined-numbered.host.path.example.com",
			path: "/testing/1",
		},
		expected: "https://first-record.host.path.example.com/testing/1",
		comment:  "for: ci-test.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains two slashes for predefined regex",
		args: data{
			host: "predefined-numbered.host.path.example.com",
			path: "/testing/",
		},
		expected: "https://second-record.host.path.example.com/testing/",
		comment:  "for: ci-test.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path is one part for predefined regex",
		args: data{
			host: "predefined-numbered.host.path.example.com",
			path: "/testing",
		},
		expected: "https://third-record.host.path.example.com/testing",
		comment:  "for: ci-test.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path is a semantic version",
		args: data{
			host: "predefined-multinumbered.host.path.example.com",
			path: "/v0.0.1/",
		},
		expected: "https://first-record.host.path.example.com/v0.0.1/",
		comment:  "for: dl.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains a word from custom regex",
		args: data{
			host: "predefined-multinumbered.host.path.example.com",
			path: "/ci-cross/test",
		},
		expected: "https://second-record.host.path.example.com/ci-cross/test",
		comment:  "for: dl.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains a word from custom regex",
		args: data{
			host: "predefined-multinumbered.host.path.example.com",
			path: "/ci/test",
		},
		expected: "https://second-record.host.path.example.com/ci/test",
		comment:  "for: dl.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains any character",
		args: data{
			host: "predefined-multinumbered.host.path.example.com",
			path: "/notdefined",
		},
		expected: "https://third-record.host.path.example.com/notdefined",
		comment:  "for: dl.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains a version",
		args: data{
			host: "predefined-version.host.path.example.com",
			path: "/v0.0/beta",
		},
		expected: "https://first-record.host.path.example.com/v0.0/beta",
		comment:  "for: docs.k8s.io",
	},
	{
		name: "Redirect to second host record's to= when path doesn't contain a version",
		args: data{
			host: "predefined-version.host.path.example.com",
			path: "/testing",
		},
		expected: "https://second-record.host.path.example.com/docs/testing",
		comment:  "for: docs.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains a version and word",
		args: data{
			host: "predefined-versionword.host.path.example.com",
			path: "/v0.0/beta",
		},
		expected: "https://first-record.host.path.example.com/release-0.0/examples/beta",
		comment:  "for: examples.k8s.io",
	},
	{
		name: "Redirect to second host record's to= when path doesn't contain a version and word",
		args: data{
			host: "predefined-versionword.host.path.example.com",
			path: "/testing",
		},
		expected: "https://second-record.host.path.example.com/testing",
		comment:  "for: examples.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains brackets",
		args: data{
			host: "predefined-simpletospecific.host.path.example.com",
			path: "/[test]/",
		},
		expected: "https://first-record.host.path.example.com/[test]/",
		comment:  "for: git.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains brackets and words",
		args: data{
			host: "predefined-simpletospecific.host.path.example.com",
			path: "/[test]/testing",
		},
		expected: "https://first-record.host.path.example.com/[test]/testing",
		comment:  "for: git.k8s.io",
	},
	{
		name: "Redirect to second host record's to= when path doesn't contain brackets",
		args: data{
			host: "predefined-simpletospecific.host.path.example.com",
			path: "/testing",
		},
		expected: "https://second-record.host.path.example.com/testing",
		comment:  "for: git.k8s.io",
	},
	{
		name: "Redirect to host record's to= using multiple predefined regexes",
		args: data{
			host: "path-subdomain.host.path.example.com",
			path: "/good-first-issue",
		},
		expected: "https://path-subdomain-redirect.host.path.example.com/4",
		comment:  "for: go.k8s.io",
	},
	{
		name: "Redirect to host record's to= when path contains release",
		args: data{
			host: "predefined-release.host.path.example.com",
			path: "/v0.0.1/beta",
		},
		expected: "https://predefined-regex.host.path.example.com/v0.0.1/beta",
		comment:  "for: releases.k8s.io",
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
