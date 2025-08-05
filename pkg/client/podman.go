package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/containerd/errdefs"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/silenium-dev/docker-wrapper/pkg/client/podman"
)

var ErrNotPodman = fmt.Errorf("not a podman server")

func (c *Client) GetPodmanConnection(ctx context.Context) (context.Context, error) {
	if ok, err := c.IsPodman(ctx); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNotPodman
	}

	cliHost := c.DaemonHost()
	connCtx, err := bindings.NewConnection(ctx, cliHost)
	if err == nil {
		return connCtx, nil
	}

	dest, err := podman.GetDestination()
	if errdefs.IsNotFound(err) && runtime.GOOS == "linux" {
		uid := os.Getuid()
		var err1, err2 error
		if uid != 0 {
			connCtx, err1 = bindings.NewConnection(ctx, fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", uid))
		}
		if err1 != nil || uid == 0 {
			connCtx, err2 = bindings.NewConnection(ctx, "unix:///run/podman/podman.sock")
		}
		if (err1 != nil || uid == 0) && err2 != nil {
			return nil, fmt.Errorf("probably not podman, or sockets are not accessible: %w", errors.Join(err1, err2))
		}
	} else if err != nil {
		return nil, err
	} else {
		connCtx, err = bindings.NewConnectionWithIdentity(ctx, dest.URI, dest.Identity, dest.IsMachine)
		if err != nil {
			return nil, err
		}
	}
	return connCtx, nil
}

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
