package client

import (
	"context"
	"fmt"

	"github.com/containers/common/pkg/config"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/system"
	errdefs2 "github.com/docker/docker/errdefs"
)

var ErrNotPodman = fmt.Errorf("not a podman server")
var ErrSocketNotEnabled = fmt.Errorf("podman socket is not enabled")

// PodmanSocket returns the path to the podman socket on the podman host.
// When talking with a podman machine, this might not be reachable for the caller.
func (c *Client) PodmanSocket() (string, error) {
	connCtx, err := c.GetPodmanConnection(context.Background())
	if err != nil {
		return "", err
	}
	info, err := system.Info(connCtx, &system.InfoOptions{})
	if err != nil {
		return "", err
	}
	if info.Host == nil || info.Host.RemoteSocket == nil || !info.Host.RemoteSocket.Exists {
		return "", ErrSocketNotEnabled
	}
	return info.Host.RemoteSocket.Path, nil
}

// GetPodmanConnection returns a connection context to the podman host.
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

	dest, err := c.getConnection()
	if err != nil {
		return nil, err
	}
	return bindings.NewConnectionWithIdentity(ctx, dest.URI, dest.Identity, dest.IsMachine)
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

func (c *Client) getDefaultConnection(conf *config.Config) (config.Connection, error) {
	conns, err := conf.GetAllConnections()
	if err != nil {
		return config.Connection{}, err
	}
	for _, conn := range conns {
		if conn.Default {
			return conn, nil
		}
	}
	return config.Connection{}, errdefs2.NotFound(fmt.Errorf("no default connection found"))
}
