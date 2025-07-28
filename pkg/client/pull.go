package client

import (
	"context"
	"fmt"
	"github.com/containerd/platforms"
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
	v2 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
)

func (c *Client) ImagePullWithEvents(ctx context.Context, ref reference.Named, options image.PullOptions) (chan events.PullEvent, error) {
	var encodedAuth string
	var err error
	if c.authProvider != nil {
		c.logger.Debugf("using configured auth provider")
		encodedAuth, err = registry.EncodeAuthConfig(c.authProvider.AuthConfig(ref))
		if err != nil {
			return nil, err
		}
	}
	options.RegistryAuth = encodedAuth
	reader, err := c.Client.ImagePull(ctx, ref.String(), options)
	if err != nil {
		return nil, err
	}
	return pull.ParseStream(ctx, reader), nil
}

func (c *Client) ImagePullWithState(ctx context.Context, ref reference.Named, options image.PullOptions) (chan state.Pull, error) {
	var platform *v1.Platform
	var err error
	if options.Platform != "" {
		platform, err = v1.ParsePlatform(options.Platform)
		if err != nil {
			return nil, err
		}
	}
	manifest, err := c.ImageGetManifest(ctx, ref, platform)
	if err != nil {
		return nil, err
	}
	eventChan, err := c.ImagePullWithEvents(ctx, ref, options)
	if err != nil {
		return nil, err
	}
	return pull.StateFromStream(ctx, ref, eventChan, manifest), nil
}

func (c *Client) ImagePull(ctx context.Context, ref reference.Named, options image.PullOptions) (digest.Digest, error) {
	eventChan, err := c.ImagePullWithEvents(ctx, ref, options)
	if err != nil {
		return "", err
	}

	var digestEvent *events.Digest
	for event := range eventChan {
		if ev, ok := event.(*events.Digest); digestEvent == nil && ok {
			c.logger.Debugf("received digest event: %s", ev.String())
			digestEvent = ev
		}
	}
	if digestEvent == nil {
		return "", fmt.Errorf("no digest event received")
	}

	return digestEvent.Digest, nil
}

func (c *Client) ImageGetManifest(ctx context.Context, ref reference.Named, platform *v1.Platform) (*v1.Manifest, error) {
	var err error
	if platform == nil {
		platform, err = c.ImageDefaultPlatform(ctx)
		if err != nil {
			return nil, err
		}
	}
	opts := []remote.Option{
		remote.WithAuthFromKeychain(c.authProvider),
		remote.WithContext(ctx),
		remote.WithPlatform(*platform),
	}

	nameRef, err := name.ParseReference(ref.String())
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(nameRef, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get image manifest: %w", err)
	}
	return img.Manifest()
}

func (c *Client) ImageDefaultPlatform(ctx context.Context) (*v1.Platform, error) {
	info, err := c.Client.Info(ctx)
	if err != nil {
		return nil, err
	}
	normalized := platforms.Normalize(v2.Platform{
		OS:           info.OSType,
		Architecture: info.Architecture,
	})

	return &v1.Platform{
		OS:           normalized.OS,
		Architecture: normalized.Architecture,
		OSVersion:    normalized.OSVersion,
		Variant:      normalized.Variant,
		OSFeatures:   normalized.OSFeatures,
	}, nil
}
