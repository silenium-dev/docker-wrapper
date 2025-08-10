package main

import (
	"bufio"
	"context"
	"os"
	"testing/fstest"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/resources"
	"k8s.io/utils/ptr"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, err := client.NewWithOpts(client.FromEnv, client.WithVersionNegotiation)
	if err != nil {
		panic(err)
	}
	res, _ := resources.NewClient(cli)
	_, err = res.GetOrCreateLoggingStack(
		ctx, map[string]*string{
			"alloy-scrape": nil,
		},
		true,
	)
	if err != nil {
		panic(err)
	}

	vol1, err := res.CreateVolume(ctx, nil, "test_volume_1")
	if err != nil {
		panic(err)
	}
	defer vol1.Remove(ctx)
	vol2, err := res.CreateVolume(ctx, nil, "test_volume_2")
	if err != nil {
		panic(err)
	}
	defer vol2.Remove(ctx)
	err = vol1.AddFiles(
		ctx, map[string]*fstest.MapFile{
			"test.txt": {
				Data: []byte("vol1-content"),
				Mode: 0o644,
			},
		}, container.CopyToContainerOptions{},
	)
	if err != nil {
		panic(err)
	}
	err = vol2.AddFiles(
		ctx, map[string]*fstest.MapFile{
			"test.txt": {
				Data: []byte("vol2-content"),
				Mode: 0o644,
			},
		}, container.CopyToContainerOptions{},
	)
	if err != nil {
		panic(err)
	}

	net1, err := res.CreateNetwork(ctx, nil, "test_network_1", network.CreateOptions{})
	if err != nil {
		panic(err)
	}
	defer net1.Remove(ctx)
	net2, err := res.CreateNetwork(ctx, nil, "test_network_2", network.CreateOptions{})
	if err != nil {
		panic(err)
	}
	defer net2.Remove(ctx)

	cont, err := res.CreateContainer(
		ctx, "test", nil, resources.ContainerSpec{
			Image: "alpine",
			Entrypoint: []string{
				"sh", "-c",
				"trap : TERM INT; while true; do cat /mnt/vol1/test.txt; echo; sleep 1; cat /mnt/vol2/test.txt; echo; sleep 1; done & wait",
			},
			Labels: map[string]string{
				"alloy-scrape": "true",
				"test":         "test-value",
				"category":     "development",
			},
			Networks: map[*resources.Network]network.EndpointSettings{
				net1: {},
				net2: {},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: vol1.Name(),
					Target: "/mnt/vol1",
				},
				{
					Type:   mount.TypeVolume,
					Source: vol2.Name(),
					Target: "/mnt/vol2",
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	println(cont.Id())
	err = cont.Start(ctx, container.StartOptions{})
	if err != nil {
		panic(err)
	}
	println("Container started")

	endpoint1, err := cont.GetEndpoint(ctx, net1)
	if err != nil {
		panic(err)
	}
	println("net1 IP:", endpoint1.IPAddress)
	endpoint2, err := cont.GetEndpoint(ctx, net2)
	if err != nil {
		panic(err)
	}
	println("net2 IP:", endpoint2.IPAddress)

	println("Press enter to stop and remove the container")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	err = cont.Stop(ctx, container.StopOptions{Timeout: ptr.To(1)})
	if err != nil {
		panic(err)
	}
	println("Container stopped")
	err = cont.Remove(ctx, container.RemoveOptions{})
	if err != nil {
		panic(err)
	}
	println("Container removed")
}
