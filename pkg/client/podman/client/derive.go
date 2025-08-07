package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/blang/semver/v4"
	"github.com/containers/podman/v5/pkg/bindings"
	errdefs2 "github.com/docker/docker/errdefs"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	config2 "github.com/silenium-dev/docker-wrapper/pkg/client/podman/config"
	"go.uber.org/zap"
)

var ErrNotPodman = fmt.Errorf("not a podman server")

func getPodmanConnection(cli *client.Client, ctx context.Context, logger *zap.SugaredLogger) (*bindings.Connection, *semver.Version, error) {
	if ok, err := cli.IsPodman(ctx); err != nil {
		return nil, nil, err
	} else if !ok {
		return nil, nil, ErrNotPodman
	}
	logger.Debugf("remote is podman")

	cliHost := cli.DaemonHost()
	logger.Debugf("trying to connect directly to docker host: %s", cliHost)
	_url, err := url.Parse(cliHost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse docker host: %w", err)
	}
	conn, ver, err := directConnection(ctx, _url)
	if err != nil {
		logger.Debugf("trying to resolve podman connection")
		dest, err := getConnection(ctx, cli, logger)
		if err != nil {
			return nil, nil, err
		}
		connCtx, err := bindings.NewConnectionWithIdentity(ctx, dest.URI, dest.Identity, dest.IsMachine)
		if err != nil {
			return nil, nil, err
		}
		conn, _ = bindings.GetClient(connCtx)
		ver = bindings.ServiceVersion(connCtx)
	}

	return conn, ver, nil
}

func deriveConnection(ctx context.Context, cli *client.Client, conf *config2.Config, logger *zap.SugaredLogger) (config2.Connection, error) {
	conns, err := conf.GetAllConnections()
	if err != nil {
		return config2.Connection{}, err
	}
	logger.Debugf("Got %d podman connections", len(conns))

	dockerInfo, err := cli.Info(ctx)
	if err != nil {
		return config2.Connection{}, fmt.Errorf("failed to get info from docker api")
	}
	logger.Debugf("looking for podman remote with id: %s", dockerInfo.ID)

	for _, conn := range conns {
		logger.Debugf("connected to %s (%s)", conn.Name, conn.URI)
		if conn.Default {
			logger.Debugf("found default connection: %s (%s)", conn.Name, conn.URI)
			logger.Debugf("trying to connect to %s (%s)", conn.Name, conn.URI)
			_, err = bindings.NewConnectionWithIdentity(ctx, conn.URI, conn.Identity, conn.IsMachine)
			if err != nil {
				logger.Errorf("failed to connect to %s (%s): %v", conn.Name, conn.URI, err)
				return conn, errdefs2.Unavailable(err)
			}
			return conn, nil
		}
	}
	return config2.Connection{}, errdefs2.NotFound(fmt.Errorf("no matching connection found for docker id: %s", dockerInfo.ID))
}
