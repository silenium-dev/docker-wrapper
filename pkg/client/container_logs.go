package client

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/silenium-dev/docker-wrapper/pkg/client/stream"
)

func (c *Client) StreamLogs(ctx context.Context, id string, follow bool) (*stream.MultiplexedStream, error) {
	inspect, err := c.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	reader, err := c.ContainerLogs(ctx, id, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: follow})
	if err != nil {
		return nil, err
	}
	return stream.NewMultiplexedStream(ctx, reader, reader, nil, !inspect.Config.Tty, c.Logger()), nil
}
