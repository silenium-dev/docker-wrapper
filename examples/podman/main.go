package main

import (
	"context"

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
}
