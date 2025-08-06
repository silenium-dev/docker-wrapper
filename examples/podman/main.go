package main

import (
	"context"

	"github.com/containers/podman/v5/pkg/bindings/system"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
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
		conn, err := cli.GetPodmanConnection(context.Background())
		if err != nil {
			panic(err)
		}
		info, err := system.Info(conn, &system.InfoOptions{})
		if err != nil {
			panic(err)
		}
		println("Rootless:", info.Host.Security.Rootless)

		socket, err := cli.PodmanSocket()
		if err != nil {
			panic(err)
		}
		println("Remote socket:", socket)
	}
}
