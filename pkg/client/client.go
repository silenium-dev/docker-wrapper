package client

import (
	"net"
	"sync"

	"github.com/docker/docker/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"go.uber.org/zap"
)

type Client struct {
	*client.Client
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
	c.Client = cli
	return c, nil
}
