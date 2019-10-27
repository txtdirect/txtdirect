package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.txtdirect.org/txtdirect/e2e/host"
)

var tests = map[string]func() error{
	"host": host.Run,
}

type dockerManager struct {
	ctx context.Context
	cli *client.Client
	dir string

	network       types.NetworkCreateResponse
	txtdContainer container.ContainerCreateCreatedBody
	cdContainer   container.ContainerCreateCreatedBody
}

func main() {
	d := dockerManager{}

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

		// Run the tests
		suiteName := strings.Split(directory, "/")[0]
		if err := tests[suiteName](); err != nil {
			log.Printf("[txtdirect_e2e]: <%s>: %s", suiteName, err)
		}

		if err := d.StopContainers(); err != nil {
			log.Fatalf("[txtdirect_e2e]: Couldn't stop containers: %s", err.Error())
		}
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
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("Couldn't start the Docker client: %s", err.Error())
	}
	cli.NegotiateAPIVersion(d.ctx)
	d.cli = cli
	return nil
}

// StartContainers starts a CoreDNS and a TXTDirect containers for each test-case
func (d *dockerManager) StartContainers() error {
	// Get current working directory to create mount test-case's data to containers
	cdir, err := os.Getwd()
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
				Source: cdir + "/" + d.dir,
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
				Source: cdir + "/" + d.dir,
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
	if err := d.cli.NetworkRemove(d.ctx, d.network.ID); err != nil {
		return fmt.Errorf("Couldn't remove the network adaptor: %s", err.Error())
	}
	return nil
}
