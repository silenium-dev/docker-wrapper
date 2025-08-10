package client

import (
	"context"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
	"go.uber.org/zap"
)

func (c *Client) ImagePullWithEvents(ctx context.Context, ref reference.Named, options image.PullOptions) (
	v1.Hash, *v1.Manifest, chan events.PullEvent, error,
) {
	if options.RegistryAuth != "" || options.PrivilegeFunc != nil {
		c.logger.WithOptions(zap.AddStacktrace(zap.DPanicLevel)).Warnf("privilege function and registry auth in options are not supported, please use auth provider instead")
		options.RegistryAuth = ""
		options.PrivilegeFunc = nil
	}
	var encodedAuth string
	var err error
	if c.authProvider != nil {
		c.logger.Debugf("using configured auth provider")
		encodedAuth, err = registry.EncodeAuthConfig(c.authProvider.AuthConfig(ref))
		if err != nil {
			return v1.Hash{}, nil, nil, err
		}
	}
	options.RegistryAuth = encodedAuth

	imageId, manifest, err := c.getManifest(ctx, ref, options)
	if err != nil {
		return v1.Hash{}, nil, nil, err
	}

	reader, err := c.APIClient.ImagePull(ctx, ref.String(), options)
	if err != nil {
		return v1.Hash{}, nil, nil, err
	}

	return imageId, manifest, pull.ParseStream(ctx, reader), nil
}

func (c *Client) ImagePullWithState(ctx context.Context, ref reference.Named, options image.PullOptions) (
	v1.Hash, *v1.Manifest, chan state.Pull, error,
) {
	isPodman, err := c.SystemIsPodman(ctx)
	if err != nil {
		return v1.Hash{}, nil, nil, err
	}

	id, manifest, eventChan, err := c.ImagePullWithEvents(ctx, ref, options)
	if err != nil {
		return v1.Hash{}, nil, nil, err
	}

	return id, manifest, pull.StateFromStream(ctx, ref, isPodman, eventChan, manifest, id), nil
}

func (c *Client) ImagePullSimple(ctx context.Context, ref reference.Named, options image.PullOptions) (
	digest.Digest, error,
) {
	dig, _, eventChan, err := c.ImagePullWithEvents(ctx, ref, options)
	if err != nil {
		return "", err
	}

	for range eventChan {
	}

	return digest.Digest(dig.String()), nil
}

func (c *Client) getManifest(ctx context.Context, ref reference.Named, options image.PullOptions) (
	v1.Hash, *v1.Manifest, error,
) {
	var platform *v1.Platform
	var err error
	if options.Platform != "" {
		platform, err = v1.ParsePlatform(options.Platform)
		if err != nil {
			return v1.Hash{}, nil, err
		}
	}
	return c.ImageGetManifest(ctx, ref, platform)
}
