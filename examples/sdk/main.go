package main

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/registry"
	client2 "github.com/docker/docker/client"
	"github.com/docker/go-sdk/container"
	"github.com/docker/go-sdk/image"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"github.com/silenium-dev/docker-wrapper/pkg/client/stream"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewDevelopment()).Sugar()
	authProvider, err := provider.NewDefaultAuthProvider()
	if err != nil {
		panic(err)
	}
	override := provider.NewOverridingAuthProvider(
		authProvider, map[string]registry.AuthConfig{},
		provider.WithSugaredLogger(logger.With(zap.String("component", "auth-provider"))),
	).WithOverride("quay.io", registry.AuthConfig{})
	cli, err := client.NewWithOpts(
		client.WithAuthProvider(override),
		client.FromEnv,
		client.WithVersionNegotiation,
		client.WithDockerOpts(client2.WithTimeout(time.Hour*1)),
		client.WithSugaredLogger(logger.With(zap.String("component", "docker-client"))),
	)
	if err != nil {
		panic(err)
	}
	err = image.Pull(context.Background(), "localstack/localstack:4", image.WithPullClient(cli.SdkClient()))
	if err != nil {
		panic(err)
	}
	cont, err := container.Run(
		context.Background(),
		container.WithName("test"),
		container.WithImage("localstack/localstack:4"),
		container.WithNoStart(),
	)
	if err != nil {
		panic(err)
	}
	err = cont.Start(context.Background())
	if err != nil {
		panic(err)
	}
	exitCode, reader, err := cont.Exec(context.Background(), []string{"echo", "hello world"})
	if err != nil {
		panic(err)
	}
	println("exit-code", exitCode)
	multiplexed := stream.NewMultiplexedStream(context.Background(), reader, nil, nil, true, logger)
	done := false
	for !done {
		select {
		case msg, ok := <-multiplexed.Messages():
			if ok {
				print(msg.StreamType.Name(), ": ", string(msg.Content))
			}
		case <-multiplexed.Done():
			done = true
			println("done")
		}
	}
	err = cont.Terminate(context.Background(), container.TerminateTimeout(1*time.Second))
	if err != nil {
		panic(err)
	}
}
