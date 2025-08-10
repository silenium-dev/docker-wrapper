package client

import (
	"context"
	"io"
	"maps"

	"github.com/docker/docker/api/types/build"
)

func (c *Client) ImageBuild(
	ctx context.Context, buildContext io.Reader, opts build.ImageBuildOptions,
) (build.ImageBuildResponse, error) {
	authConfigs := c.authProvider.AuthConfigs()
	maps.Copy(authConfigs, opts.AuthConfigs)
	opts.AuthConfigs = authConfigs

	return c.APIClient.ImageBuild(ctx, buildContext, opts)
}
