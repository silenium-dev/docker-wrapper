package client

import (
	"net/http"
	"slices"

	"github.com/docker/docker/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/auth"
	"go.uber.org/zap"
)

type Client struct {
	*client.Client
	dockerOpts   []client.Opt
	authProvider auth.Provider
	logger       *zap.SugaredLogger
}

type Opt func(*Client) error

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
	cli, err := client.NewClientWithOpts(c.dockerOpts...)
	if err != nil {
		return nil, err
	}
	c.Client = cli
	return c, nil
}

func (c *Client) Close() error {
	return c.Client.Close()
}

func WithVersionNegotiation(c *Client) error {
	c.dockerOpts = append(c.dockerOpts, client.WithAPIVersionNegotiation())
	return nil
}

func WithAuthProvider(authProvider auth.Provider) Opt {
	return func(c *Client) error {
		c.authProvider = authProvider
		return nil
	}
}

func WithSugaredLogger(logger *zap.SugaredLogger) Opt {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

func WithLogger(logger *zap.Logger) Opt {
	return func(c *Client) error {
		c.logger = logger.Sugar()
		return nil
	}
}

func WithDockerOpts(opts ...client.Opt) Opt {
	return func(c *Client) error {
		c.dockerOpts = slices.Concat(c.dockerOpts, opts)
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) Opt {
	return func(c *Client) error {
		c.dockerOpts = append(c.dockerOpts, client.WithHTTPClient(httpClient))
		return nil
	}
}

func FromEnv(c *Client) error {
	c.dockerOpts = append(c.dockerOpts, client.FromEnv)
	return nil
}
