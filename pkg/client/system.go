package client

import (
	"context"

	"github.com/containerd/platforms"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	v2 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (c *Client) SystemIsPodman(ctx context.Context) (bool, error) {
	ver, err := c.ServerVersion(ctx)
	if err != nil {
		return false, err
	}
	for _, c := range ver.Components {
		if c.Name == "Podman Engine" {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) SystemDefaultPlatform(ctx context.Context) (*v1.Platform, error) {
	info, err := c.APIClient.Info(ctx)
	if err != nil {
		return nil, err
	}
	normalized := platforms.Normalize(
		v2.Platform{
			OS:           info.OSType,
			Architecture: info.Architecture,
		},
	)

	return &v1.Platform{
		OS:           normalized.OS,
		Architecture: normalized.Architecture,
		OSVersion:    normalized.OSVersion,
		Variant:      normalized.Variant,
		OSFeatures:   normalized.OSFeatures,
	}, nil
}
