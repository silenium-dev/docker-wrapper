package client

import (
	"context"
	"github.com/docker/docker/api/types"
	"io"
	"maps"
)

func (c *Client) ImageBuild(ctx context.Context, buildContext io.Reader, opts types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	authConfigs := c.authProvider.AuthConfigs()
	maps.Copy(authConfigs, opts.AuthConfigs)
	opts.AuthConfigs = authConfigs

	return c.Client.ImageBuild(ctx, buildContext, opts)
}
