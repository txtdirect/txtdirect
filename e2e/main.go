package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	corednsImage = "k8s.gcr.io/coredns:1.6.2"
	testerImage  = "c.txtdirect.org/tester:dirty"
)

var txtdirectImage = fmt.Sprintf("c.txtdirect.org/txtdirect:%s", os.Getenv("VERSION"))

var resultRegex = regexp.MustCompile("Total:+\\s(\\d+),\\sPassed:+\\s(\\d+),\\sFailed:+\\s(\\d+)")

type dockerManager struct {
	dir       string
	cdir      string
	gomodpath string

	testerContainerLogs map[string][]byte

	stats struct {
		total  int
		passed int
		failed int
	}
}

func main() {
	d := dockerManager{
		testerContainerLogs: make(map[string][]byte),
	}

	if err := d.CheckClient(); err != nil {
		log.Fatalf("[txtdirect_e2e]: Docker daemon didn't respond to client: %s", err)
	}

	var directories []string
	if err := listDirectories(&directories); err != nil {
		log.Printf("[txtdirect_e2e]: Couldn't list the test directories: %s", err.Error())
	}

	if err := d.CreateNetwork(); err != nil {
		log.Fatalf("[txtdirect_e2e]: Couldn't create the network adaptor: %s", err.Error())
	}

	// Run the tests for each test-case
	for _, directory := range directories {
		// Start the CoreDNS and TXTDirect containers for test-case
		d.dir = directory
		if err := d.StartContainers(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't start containers: %s", err.Error())
		}

		if err := d.RunTesterContainer(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't run the tests in container: %s", err.Error())
		}

		if err := d.StopContainers(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't stop containers: %s", err.Error())
		}
	}

	if err := d.ExamineLogs(); err != nil {
		log.Fatalf("[txtdirect_e2e]: Couldn't examine the logs and count the stats: %s", err.Error())
	}

	if err := d.RemoveNetwork(); err != nil {
		log.Fatalf("[txtdirect_e2e]: Couldn't remove the network adaptor: %s", err.Error())
	}

	if d.stats.failed != 0 {
		log.Fatalf("[txtdirect_e2e]: Tests failed because there were %d failed tests.", d.stats.failed)
	}
}

func listDirectories(directories *[]string) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Couldn't find the directory %q: %s", path, err.Error())
		}
		if info.IsDir() && path != "." {
			*directories = append(*directories, path)
			return filepath.SkipDir
		}
		return nil
	})
}

// CheckClient calls the "docker info" command to check if Docker daemon and CLI exist
func (d *dockerManager) CheckClient() error {
	_, err := exec.Command("docker", "info").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?")
	}

	return nil
}

// StartContainers starts a CoreDNS and a TXTDirect container for each test-case
func (d *dockerManager) StartContainers() error {
	// Get current working directory to create mountpoint to containers for test-case data
	var err error
	d.cdir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Couldn't get the current working directory: %s", err.Error())
	}

	// Fetch GOPATH for mounting /go/pkg/mod volumes
	if os.Getenv("GOPATH") == "" {
		return fmt.Errorf("$GOPATH is empty")
	}
	d.gomodpath = fmt.Sprintf("%s/pkg/mod", os.Getenv("GOPATH"))

	// Create the CoreDNS container
	_, err = exec.Command("docker",
		"container", "run",
		"-d",
		// Exposed Ports
		"-p", "53:53",
		// Mounted volumes
		"-v", fmt.Sprintf("%s/%s:/e2e", d.cdir, d.dir),
		"-v", fmt.Sprintf("%s:/go/pkg/mod", d.gomodpath),
		// Network
		"--network", "coretxtd",
		// Container Name
		"--name", "e2e_coredns_container",
		corednsImage,
		// CMD
		"-conf",
		"/e2e/Corefile",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't create the CoreDNS container: %s", err.Error())
	}

	// Create the TXTDirect container
	_, err = exec.Command("docker",
		"container", "run",
		"-d",
		// Exposed Ports
		"-p", "80:80",
		// Mounted volumes
		"-v", fmt.Sprintf("%s/%s:/e2e", d.cdir, d.dir),
		"-v", fmt.Sprintf("%s:/go/pkg/mod", d.gomodpath),
		// Network
		"--network", "coretxtd",
		// Container Name
		"--name", "e2e_txtdirect_container",
		txtdirectImage,
		// CMD
		"-conf",
		"/e2e/txtdirect.config",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't create the TXTDirect container: %s", err.Error())
	}

	// Connect the Docker registry to the e2e network
	if strings.Contains(d.dir, "dockerv2") {
		_, err := exec.Command("docker", "network", "connect", "coretxtd", "registry").CombinedOutput()
		if err != nil {
			return fmt.Errorf("Couldn't connect Docker registry the network adaptor: %s", err.Error())
		}

		// Tag TXTDirect's image to use in custom registry
		_, err = exec.Command("docker", "tag", txtdirectImage, "172.20.10.3:5000/txtdirect").CombinedOutput()
		if err != nil {
			return fmt.Errorf("Couldn't tag image to use in custom Docker registry: %s", err.Error())
		}

		_, err = exec.Command("docker", "save", "172.20.10.3:5000/txtdirect:latest", "-o", "txtdirect.tar").CombinedOutput()
		if err != nil {
			return fmt.Errorf("Couldn't export the image into a tarball: %s", err.Error())
		}

		// Push the TXTDirect image to the custom registry
		_, err = exec.Command("crane", "push", "txtdirect.tar", "172.20.10.3:5000/txtdirect").CombinedOutput()
		if err != nil {
			return fmt.Errorf("Couldn't push the image to the custom Docker registry: %s", err.Error())
		}

		_, err = exec.Command("rm", "./txtdirect.tar").CombinedOutput()
		if err != nil {
			return fmt.Errorf("Couldn't remove the image tarball: %s", err.Error())
		}
	}

	return nil
}

func (d *dockerManager) StopContainers() error {
	if strings.Contains(d.dir, "dockerv2") {
		_, err := exec.Command("docker",
			"logs",
			"e2e_txtdirect_container",
		).CombinedOutput()
		if err != nil {
			return err
		}
	}
	_, err := exec.Command("docker",
		"container", "rm", "-f",
		"e2e_coredns_container",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't remove the CoreDNS container: %s", err.Error())
	}

	_, err = exec.Command("docker",
		"container", "rm", "-f",
		"e2e_txtdirect_container",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't remove the TXTDirect container: %s", err.Error())
	}

	_, err = exec.Command("docker",
		"container", "rm", "-f",
		fmt.Sprintf("e2e_%s_container", d.dir),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't remove the tester container: %s", err.Error())
	}

	if strings.Contains(d.dir, "dockerv2") {
		_, err = exec.Command("docker",
			"network", "disconnect",
			"coretxtd",
			"registry",
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("Couldn't disconnect the Docker registry container: %s", err.Error())
		}
	}

	return nil
}

func (d *dockerManager) RunTesterContainer() error {
	_, err := user.Current()
	if err != nil {
		return fmt.Errorf("Couldn't get the current user: %s", err.Error())
	}

	// Create the tester container
	d.testerContainerLogs[d.dir], err = exec.Command("docker",
		"container", "run",
		// Mounted volumes
		"-v", fmt.Sprintf("%s/%s:/e2e", d.cdir, d.dir),
		"-v", fmt.Sprintf("%s:/go/pkg/mod", d.gomodpath),
		// DNS
		"--dns", "172.20.10.1",
		// Network
		"--network", "coretxtd",
		// Container Name
		"--name", fmt.Sprintf("e2e_%s_container", d.dir),
		testerImage,
		// CMD
		"go", "run", "main.go",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't create the tester container: %s", err.Error())
	}

	return nil
}

// ExamineLogs reads each tester container's logs and counts the failed and passed tests
func (d *dockerManager) ExamineLogs() error {
	for testCase, logs := range d.testerContainerLogs {
		log.Println(string(logs))
		// Stats slice order: [0: Full Log, 1: Total, 2: Passed, 3: Failed]
		stats := resultRegex.FindAllStringSubmatch(string(logs), -1)[0]
		if err := d.CountStats(stats); err != nil {
			return fmt.Errorf("Couldn't count the stats for \"%s\" test-case: %s", testCase, err.Error())
		}
	}
	log.Printf("Total Tests: %d, Total Passed Tests: %d, Total Failed Tests: %d", d.stats.total, d.stats.passed, d.stats.failed)
	return nil
}

func (d *dockerManager) CountStats(stats []string) error {
	total, err := strconv.Atoi(stats[1])
	if err != nil {
		return fmt.Errorf("Couldn't conver count of tests to int: %s", err.Error())
	}
	d.stats.total += total

	passed, err := strconv.Atoi(stats[2])
	if err != nil {
		return fmt.Errorf("Couldn't convert count of passed tests to int: %s", err.Error())
	}
	d.stats.passed += passed

	failed, err := strconv.Atoi(stats[3])
	if err != nil {
		return fmt.Errorf("Couldn't convert count of failed tests to int: %s", err.Error())
	}
	d.stats.failed += failed

	return nil
}

func (d *dockerManager) RemoveNetwork() error {
	_, err := exec.Command("docker",
		"network", "rm",
		"coretxtd",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't remove the network adaptor: %s", err.Error())
	}
	return nil
}

func (d *dockerManager) CreateNetwork() error {
	// Create a Docker network for containers
	_, err := exec.Command("docker", "network", "create", "--ip-range=\"172.20.10.0/24\"", "--subnet=\"172.20.0.0/16\"", "coretxtd").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Couldn't create the network adaptor: %s", err.Error())
	}
	return nil
}
