package client

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/auth"
	"go.uber.org/zap"
)

type Podman struct {
	cli          *client.Client
	conn         *bindings.Connection
	ver          *semver.Version
	logger       *zap.SugaredLogger
	authProvider auth.Provider
}

// FromDocker derives a podman connection from the docker remote. Fails if remote is not a podman engine
func FromDocker(
	ctx context.Context,
	cli *client.Client,
	authProvider auth.Provider,
	logger *zap.SugaredLogger,
) (*Podman, error) {
	if logger == nil {
		_logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		logger = _logger.Sugar()
	}
	conn, ver, err := getPodmanConnection(cli, ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve podman connection: %w", err)
	}
	return &Podman{
		cli:          cli,
		conn:         conn,
		ver:          ver,
		logger:       logger,
		authProvider: authProvider,
	}, nil
}

func mustGetClient(ctx context.Context) *bindings.Connection {
	conn, err := bindings.GetClient(ctx)
	if err != nil {
		panic(fmt.Errorf("connection missing (this is a bug!!): %w", err))
	}
	return conn
}
