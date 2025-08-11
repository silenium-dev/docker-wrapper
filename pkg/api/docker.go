package api

import (
	"context"

	"github.com/docker/docker/client"
)

type DockerClient interface {
	client.APIClient
	NewVersionError(ctx context.Context, APIrequired, feature string) error
}
