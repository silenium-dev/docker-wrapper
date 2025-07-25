package client

import (
	"context"
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
)

func (c *Client) Pull(ctx context.Context, ref reference.Named) (chan events.PullEvent, error) {
	var encodedAuth string
	var err error
	if c.authProvider != nil {
		encodedAuth, err = registry.EncodeAuthConfig(c.authProvider.AuthConfig(ref))
		if err != nil {
			return nil, err
		}
	}
	reader, err := c.Client.ImagePull(ctx, ref.String(), types.ImagePullOptions{RegistryAuth: encodedAuth})
	if err != nil {
		return nil, err
	}
	return pull.ParseStream(ctx, reader), nil
}
