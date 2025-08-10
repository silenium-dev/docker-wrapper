package client

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/silenium-dev/docker-wrapper/pkg/api"
	"github.com/silenium-dev/docker-wrapper/pkg/client/podman/containers/bindings"
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"go.uber.org/zap"
)

type Podman struct {
	cli          api.ClientWrapper
	conn         *bindings.Connection
	ver          *semver.Version
	logger       *zap.SugaredLogger
	authProvider provider.AuthProvider
}

// FromDocker derives a podman connection from the docker remote. Fails if remote is not a podman engine
func FromDocker(
	ctx context.Context,
	cli api.ClientWrapper,
) (*Podman, error) {
	conn, ver, err := getPodmanConnection(cli, ctx, cli.Logger())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve podman connection: %w", err)
	}
	return &Podman{
		cli:          cli,
		conn:         conn,
		ver:          ver,
		logger:       cli.Logger(),
		authProvider: cli.AuthProvider(),
	}, nil
}

func (p *Podman) AuthProvider() provider.AuthProvider {
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
