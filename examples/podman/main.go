package main

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	client2 "github.com/silenium-dev/docker-wrapper/pkg/client/podman/client"
)

func main() {
	cli, err := client.NewWithOpts(client.FromEnv, client.WithVersionNegotiation)
	if err != nil {
		panic(err)
	}
	isPodman, err := cli.IsPodman(context.Background())
	if err != nil {
		panic(err)
	}
	println("Is Podman:", isPodman)
	if isPodman {
		podmanCli, err := client2.FromDocker(context.Background(), cli, nil, nil)
		if err != nil {
			panic(err)
		}

		info, err := podmanCli.SystemInfo(context.Background())
		if err != nil {
			panic(err)
		}
		println("Rootless:", info.Host.Security.Rootless)

		socket, err := podmanCli.RemoteSocket(context.Background())
		if err != nil {
			panic(err)
		}
		println("Remote socket:", socket)
	}

	netResp, err := cli.NetworkCreate(context.Background(), "test_network", network.CreateOptions{})
	if err != nil {
		panic(err)
	}
	defer cli.NetworkRemove(context.Background(), netResp.ID)

	hostIP, err := cli.HostIPFromContainers(context.Background(), &netResp.ID)
	if err != nil {
		panic(err)
	}
	println("Host IP from containers:", hostIP.String())
}
