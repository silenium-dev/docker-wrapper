package main

import (
	"context"
	"fmt"
	"github.com/distribution/reference"
	client2 "github.com/docker/docker/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/auth"
	"time"
)

func main() {
	authProvider, err := auth.NewDefaultProvider()
	if err != nil {
		panic(err)
	}
	cli, err := client.NewWithOpts(
		client.WithAuthProvider(authProvider),
		client.FromEnv,
		client.WithDockerOpts(client2.WithTimeout(time.Hour*1)),
	)
	if err != nil {
		panic(err)
	}
	//ref, err := reference.ParseDockerRef("quay.io/prometheus/node-exporter@sha256:a25fbdaa3e4d03e0d735fd03f231e9a48332ecf40ca209b2f103b1f970d1cde0")
	ref, err := reference.ParseDockerRef("localstack/localstack:latest")
	if err != nil {
		panic(err)
	}
	stateChan, err := cli.PullWithState(context.Background(), ref)
	if err != nil {
		panic(err)
	}
	for state := range stateChan {
		print("\033[2J")
		fmt.Printf("%s\n", state.Status())
		for idx, l := range state.Layers() {
			fmt.Printf("%02d [%s]: %s\n", idx, l.Id(), l.Status())
		}
	}
	digest, err := cli.Pull(context.Background(), ref)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Digest: %s\n", digest)
}
