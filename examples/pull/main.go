package main

import (
	"context"
	"time"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	client2 "github.com/docker/docker/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/auth"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewDevelopment()).Sugar()
	authProvider, err := auth.NewDefaultProvider()
	if err != nil {
		panic(err)
	}
	override := auth.NewOverridingProvider(
		authProvider, map[string]registry.AuthConfig{},
		auth.WithSugaredLogger(logger.With(zap.String("component", "auth-provider"))),
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
	//ref, err := reference.ParseDockerRef("quay.io/prometheus/prometheus:latest")
	ref, err := reference.ParseDockerRef("localstack/localstack:4")
	if err != nil {
		panic(err)
	}
	id, manifest, stateChan, err := cli.ImagePullWithState(context.Background(), ref, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	logger.Infof("Image ID: %s\n", id.String())
	logger.Infof("Manifest digest: %s\n", manifest.Config.Digest.String())
	for state := range stateChan {
		print("\033[2J")
		logger.Infof("%s", state.Status())
		for idx, l := range state.Layers() {
			logger.Infof("%02d [%s]: %s", idx, l.Id(), l.Status())
		}
	}
	digest, err := cli.ImagePull(context.Background(), ref, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	logger.Infof("Digest: %s\n", digest.String())
}
