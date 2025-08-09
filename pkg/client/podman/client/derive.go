package client

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/podman/containers/bindings"
	"go.uber.org/zap"
)

var ErrNotPodman = fmt.Errorf("not a podman server")

func getPodmanConnection(cli *client.Client, ctx context.Context, logger *zap.SugaredLogger) (
	*bindings.Connection, *semver.Version, error,
) {
	if ok, err := cli.SystemIsPodman(ctx); err != nil {
		return nil, nil, err
	} else if !ok {
		return nil, nil, ErrNotPodman
	}
	logger.Debugf("remote is podman")

	cliHost := cli.DaemonHost()
	logger.Debugf("trying to connect directly to docker host: %s", cliHost)

	connCtx, err := bindings.NewConnection(ctx, cliHost)
	if err != nil {
		return nil, nil, err
	}
	conn, _ := bindings.GetClient(connCtx)
	ver := bindings.ServiceVersion(connCtx)

	return conn, ver, nil
}
