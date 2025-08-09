package api

import (
	"context"
	"io"
	"net"
	"net/http"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
	"github.com/silenium-dev/docker-wrapper/pkg/client/stream"
	"go.uber.org/zap"
)

type ClientBase interface {
	AuthProvider() provider.AuthProvider
	Logger() *zap.SugaredLogger
	Close() error
	RequestAuthenticate(req *http.Request, ref reference.Named) error
}

type ImageClient interface {
	ImageBuild(
		ctx context.Context, buildContext io.Reader, opts build.ImageBuildOptions,
	) (build.ImageBuildResponse, error)
	ImagePullWithEvents(ctx context.Context, ref reference.Named, options image.PullOptions) (
		v1.Hash, *v1.Manifest, chan events.PullEvent, error,
	)
	ImagePullWithState(ctx context.Context, ref reference.Named, options image.PullOptions) (
		v1.Hash, *v1.Manifest, chan state.Pull, error,
	)
	ImagePullSimple(ctx context.Context, ref reference.Named, options image.PullOptions) (digest.Digest, error)
	ImageGetManifest(ctx context.Context, ref reference.Named, platform *v1.Platform) (v1.Hash, *v1.Manifest, error)
}

type ContainerClient interface {
	StreamLogs(ctx context.Context, id string, follow bool) (*stream.MultiplexedStream, error)
}

type SystemClient interface {
	SystemHostIPFromContainers(ctx context.Context, netId *string) (net.IP, error)
	SystemIsPodman(ctx context.Context) (bool, error)
	SystemDefaultPlatform(ctx context.Context) (*v1.Platform, error)
}

type ClientWrapper interface {
	client.APIClient
	ClientBase
	ImageClient
	ContainerClient
	SystemClient
}
