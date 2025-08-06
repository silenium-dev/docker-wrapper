//go:build unix

package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/containers/common/pkg/config"
	"github.com/containers/podman/v5/pkg/bindings"
)

func (c *Client) getConnection() (config.Connection, error) {
	queryCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Default()
	if err == nil {
		defaultConn, err := c.getDefaultConnection(conf)
		if err == nil {
			return defaultConn, nil
		}
	}
	uid := os.Getuid()
	conn := config.Connection{
		Default:   true,
		Name:      "podman-rootless",
		ReadWrite: true,
		Destination: config.Destination{
			IsMachine: false,
		},
	}
	var err1, err2 error
	if uid != 0 {
		sock := fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", uid)
		_, err1 = bindings.NewConnection(queryCtx, sock)
		conn.Destination.URI = sock
	}
	if err1 != nil || uid == 0 {
		sock := "unix:///run/podman/podman.sock"
		_, err2 = bindings.NewConnection(queryCtx, sock)
		conn.Destination.URI = sock
	}
	if (err1 != nil || uid == 0) && err2 != nil {
		return config.Connection{}, fmt.Errorf(
			"probably not podman, or sockets are not accessible: %w", errors.Join(err1, err2),
		)
	}
	return conn, nil
}
