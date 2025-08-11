package client

import (
	"log/slog"
	"net"
	"sync"
	"unsafe"

	"github.com/docker/docker/client"
	client2 "github.com/docker/go-sdk/client"
	"github.com/silenium-dev/docker-wrapper/pkg/api"
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type Client struct {
	api.DockerClient
	sdkClient              *client2.Client
	dockerOpts             []client.Opt
	authProvider           provider.AuthProvider
	imageProvider          provider.ImageProvider
	logger                 *zap.SugaredLogger
	hostFromContainerAddr  net.IP
	hostFromContainerMutex sync.RWMutex
}

func NewWithOpts(opts ...Opt) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	if c.logger == nil {
		c.logger = zap.Must(zap.NewDevelopment()).Sugar()
	}
	if c.imageProvider == nil {
		c.imageProvider = provider.DefaultImageProvider()
	}
	cli, err := client.NewClientWithOpts(c.dockerOpts...)
	if err != nil {
		return nil, err
	}
	c.DockerClient = cli

	result := &client2.Client{}
	internal := (*sdkClient)(unsafe.Pointer(result))
	internal.dockerClient = cli
	internal.once.Do(func() {})
	internal.log = slog.New(zapslog.NewHandler(
		c.logger.Desugar().Core(),
		zapslog.WithCaller(true),
		zapslog.WithName("docker-sdk"),
		zapslog.AddStacktraceAt(slog.LevelError),
	))
	c.sdkClient = result

	return c, nil
}
