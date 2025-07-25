package main

import (
	"context"
	"docker-wrapper/pkg/client"
	"github.com/distribution/reference"
	client2 "github.com/docker/docker/client"
)

func main() {
	cli, err := client.New(client2.FromEnv)
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
