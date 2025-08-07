package client

import (
	"context"
)

func (c *Client) IsPodman(ctx context.Context) (bool, error) {
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
