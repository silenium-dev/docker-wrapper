package main

import (
	"context"
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
		client.WithDockerOpts(client2.WithTimeout(time.Second*10)),
	)
	if err != nil {
		panic(err)
	}
	//ref, err := reference.ParseDockerRef("quay.io/prometheus/node-exporter@sha256:a25fbdaa3e4d03e0d735fd03f231e9a48332ecf40ca209b2f103b1f970d1cde0")
	ref, err := reference.ParseDockerRef("confluentinc/cp-kafka:latest")
	if err != nil {
		panic(err)
	}
	eventChan, err := cli.Pull(context.Background(), ref)
	if err != nil {
		panic(err)
	}
	for event := range eventChan {
		println(event.String())
	}
}
