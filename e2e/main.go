package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var resultRegex = regexp.MustCompile("Total:+\\s(\\d+),\\sPassed:+\\s(\\d+),\\sFailed:+\\s(\\d+)")

type dockerManager struct {
	ctx  context.Context
	cli  *client.Client
	dir  string
	cdir string

	network             types.NetworkCreateResponse
	txtdContainer       container.ContainerCreateCreatedBody
	cdContainer         container.ContainerCreateCreatedBody
	testerContainer     container.ContainerCreateCreatedBody
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

	if err := d.CreateClient(); err != nil {
		log.Fatalf("[txtdirect_e2e]: Docker daemon didn't respond to client: %s", err)
	}

	var directories []string
	if err := listDirectories(&directories); err != nil {
		log.Printf("[txtdirect_e2e]: Couldn't list the test directories: %s", err.Error())
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

		if err := d.WaitForLogs(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't wait for the logs: %s", err.Error())
		}

		if err := d.ReadTestLogs(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't read the tests logs: %s", err.Error())
		}

		if err := d.StopContainers(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't stop containers: %s", err.Error())
		}
	}

	if err := d.ExamineLogs(); err != nil {
		log.Fatalf("[txtdirect_e2e]: Couldn't examine the logs and count the stats: %s", err.Error())
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

// CreateClient creates a Docker client and context and attaches them to the
// dockerManager instance
func (d *dockerManager) CreateClient() error {
	d.ctx = context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("Couldn't start the Docker client: %s", err.Error())
	}
	d.cli = cli
	return nil
}

// StartContainers starts a CoreDNS and a TXTDirect containers for each test-case
func (d *dockerManager) StartContainers() error {
	// Get current working directory to create mount test-case's data to containers
	var err error
	d.cdir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Couldn't get the current working directory: %s", err.Error())
	}

	d.network, err = d.cli.NetworkCreate(d.ctx, "coretxtd", types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					IPRange: "172.20.10.0/24",
					Subnet:  "172.20.0.0/16",
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("Couldn't create the network adaptor: %s", err.Error())
	}

	// Create the CoreDNS container
	d.cdContainer, err = d.cli.ContainerCreate(d.ctx, &container.Config{
		Image: "coredns/coredns",
		Cmd:   []string{"-conf", "/root/Corefile"},
		ExposedPorts: nat.PortSet{
			"53/tcp": struct{}{},
			"53/udp": struct{}{},
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: d.cdir + "/" + d.dir,
				Target: "/root",
			},
		},
		PortBindings: nat.PortMap{
			"53/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "53",
				},
			},
			"53/udp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "53",
				},
			},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"coretxtd": &network.EndpointSettings{NetworkID: d.network.ID},
		},
	}, "")
	if err != nil {
		return fmt.Errorf("Couldn't create the CoreDNS container: %s", err.Error())
	}

	d.txtdContainer, err = d.cli.ContainerCreate(d.ctx, &container.Config{
		Image: "okkur/txtdirect:0.4.0",
		Cmd:   []string{"-conf", "/root/TXTD.config"},
		ExposedPorts: nat.PortSet{
			"80/tcp": struct{}{},
			"80/udp": struct{}{},
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: d.cdir + "/" + d.dir,
				Target: "/root",
			},
		},
		PortBindings: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "80",
				},
			},
			"80/udp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "80",
				},
			},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"coretxtd": &network.EndpointSettings{NetworkID: d.network.ID},
		},
	}, "")
	if err != nil {
		return fmt.Errorf("Couldn't create the TXTDirect container: %s", err.Error())
	}

	// Start the CoreDNS container
	if err := d.cli.ContainerStart(d.ctx, d.cdContainer.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Couldn't start the CoreDNS container: %s", err.Error())
	}

	// Start the TXTDirect container
	if err := d.cli.ContainerStart(d.ctx, d.txtdContainer.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Couldn't start the TXTDirect container: %s", err.Error())
	}

	return nil
}

func (d *dockerManager) StopContainers() error {
	if err := d.cli.ContainerStop(d.ctx, d.cdContainer.ID, nil); err != nil {
		return fmt.Errorf("Couldn't remove the CoreDNS container: %s", err.Error())
	}
	if err := d.cli.ContainerStop(d.ctx, d.txtdContainer.ID, nil); err != nil {
		return fmt.Errorf("Couldn't remove the TXTDirect container: %s", err.Error())
	}
	if err := d.cli.ContainerStop(d.ctx, d.testerContainer.ID, nil); err != nil {
		return fmt.Errorf("Couldn't remove the tester container: %s", err.Error())
	}
	if err := d.cli.NetworkRemove(d.ctx, d.network.ID); err != nil {
		return fmt.Errorf("Couldn't remove the network adaptor: %s", err.Error())
	}
	return nil
}

func (d *dockerManager) RunTesterContainer() error {
	var err error
	d.testerContainer, err = d.cli.ContainerCreate(d.ctx, &container.Config{
		Image: "c.txtdirect.org/tester:0.0.1",
		Cmd:   []string{"go", "run", "main.go"},
	}, &container.HostConfig{
		DNS: []string{"172.20.10.1"},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: d.cdir + "/" + d.dir,
				Target: "/root",
			},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"coretxtd": &network.EndpointSettings{NetworkID: d.network.ID},
		},
	}, "")
	if err != nil {
		return fmt.Errorf("Couldn't create the tester container: %s", err.Error())
	}

	if err := d.cli.ContainerStart(d.ctx, d.testerContainer.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Couldn't start the tester container: %s", err.Error())
	}

	return nil
}

func (d *dockerManager) ReadTestLogs() error {
	logsReader, err := d.cli.ContainerLogs(d.ctx, d.testerContainer.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return fmt.Errorf("Couldn't get the tester container logs: %s", err.Error())
	}

	d.testerContainerLogs[d.dir], err = ioutil.ReadAll(logsReader)
	if err != nil {
		return fmt.Errorf("Couldn't read the tester container logs: %s", err.Error())
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

func (d *dockerManager) WaitForLogs() error {
	status, err := d.cli.ContainerWait(d.ctx, d.testerContainer.ID)
	if err != nil {
		return fmt.Errorf("Couldn't wait for the tester container: %s", err.Error())
	}
	if status != 0 {
		return fmt.Errorf("Wait response's status code is wrong: %#+v", status)
	}
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
