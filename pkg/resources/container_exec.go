package resources

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/silenium-dev/docker-wrapper/pkg/client/stream"
)

func (c *Container) Exec(
	ctx context.Context,
	cmd []string, env map[string]string,
	workingDir *string, user *string,
) (*stream.MultiplexedStream, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return nil, err
	}

	options := container.ExecOptions{
		Tty:          false,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Cmd:          cmd,
	}
	if workingDir != nil {
		options.WorkingDir = *workingDir
	}
	if user != nil {
		options.User = *user
	}
	for k, v := range env {
		options.Env = append(options.Env, k+"="+v)
	}

	create, err := c.client.wrapper.ContainerExecCreate(ctx, c.id, options)
	if err != nil {
		return nil, err
	}
	err = c.client.wrapper.ContainerExecStart(ctx, create.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, err
	}
	attach, err := c.client.wrapper.ContainerExecAttach(ctx, create.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, err
	}
	return stream.NewMultiplexedStream(
		ctx, attach.Reader, attach.Conn, attach.Conn, true, c.client.wrapper.Logger(),
	), nil
}
