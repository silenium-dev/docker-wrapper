package client

import (
	"context"
	"fmt"
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
)

func (c *Client) PullWithState(ctx context.Context, ref reference.Named) (chan state.Pull, error) {
	eventChan, err := c.PullWithEvents(ctx, ref)
	if err != nil {
		return nil, err
	}
	return pull.StateFromStream(ctx, ref, eventChan), nil
}

func (c *Client) PullWithEvents(ctx context.Context, ref reference.Named) (chan events.PullEvent, error) {
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

func (c *Client) Pull(ctx context.Context, ref reference.Named) (digest.Digest, error) {
	eventChan, err := c.PullWithEvents(ctx, ref)
	if err != nil {
		return "", err
	}

	var digestEvent *events.Digest
	for event := range eventChan {
		if _, ok := event.(*events.Digest); digestEvent == nil && ok {
			digestEvent = event.(*events.Digest)
		}
	}
	if digestEvent == nil {
		return "", fmt.Errorf("no digest event received")
	}

	return digestEvent.Digest, nil
}
