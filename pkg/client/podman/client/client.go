package client

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/auth"
	"github.com/silenium-dev/docker-wrapper/pkg/client/podman/containers/bindings"
	"go.uber.org/zap"
)

type Podman struct {
	cli          *client.Client
	conn         *bindings.Connection
	ver          *semver.Version
	logger       *zap.SugaredLogger
	authProvider auth.AuthProvider
}

// FromDocker derives a podman connection from the docker remote. Fails if remote is not a podman engine
func FromDocker(
	ctx context.Context,
	cli *client.Client,
	authProvider auth.AuthProvider,
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

func (p *Podman) AuthProvider() auth.AuthProvider {
	return p.authProvider
}

func (p *Podman) Connection() *bindings.Connection {
	return p.conn
}

func (p *Podman) APIVersion() *semver.Version {
	return p.ver
}

func (p *Podman) Logger() *zap.SugaredLogger {
	return p.logger
}
