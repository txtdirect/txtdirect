package main

import (
	"fmt"
	"log"
	"os/exec"
)

type data struct {
	host string
	path string
	dir  string
}

type test struct {
	name string
	args data
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
}

func main() {
	result := make(map[bool][]test)
	for _, test := range tests {
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
