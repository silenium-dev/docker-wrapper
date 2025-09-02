package client

import (
	"context"
	"fmt"

	"github.com/distribution/reference"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func (c *Client) ImageGetManifest(ctx context.Context, ref reference.Named, platform *v1.Platform) (
	v1.Hash, *v1.Manifest, error,
) {
	var err error
	if platform == nil {
		platform, err = c.SystemDefaultPlatform(ctx)
		if err != nil {
			return v1.Hash{}, nil, err
		}
	}
	opts := []remote.Option{
		remote.WithAuthFromKeychain(c.authProvider),
		remote.WithContext(ctx),
		remote.WithPlatform(*platform),
	}

	nameRef, err := name.ParseReference(ref.String())
	if err != nil {
		return v1.Hash{}, nil, err
	}

	desc, err := remote.Get(nameRef, opts...)
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("failed to get descriptor: %w", err)
	}

	img, err := desc.Image()
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("failed to get image: %w", err)
	}
	isPodman, err := c.SystemIsPodman(ctx)
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("failed to determine if runtime is podman: %w", err)
	}
	manifest, err := img.Manifest()
	if err != nil {
		return v1.Hash{}, nil, err
	}

	var id v1.Hash
	if isPodman {
		// Podman uses config digest as image-id
		id, err = img.ConfigName()
	} else {
		// Docker uses manifest/index digest as image-id
		id = desc.Digest
	}
	if err != nil {
		return v1.Hash{}, nil, err
	}

	return id, manifest, nil
}
