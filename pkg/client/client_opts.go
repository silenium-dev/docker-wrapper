package client

import (
	"net/http"
	"slices"

	"github.com/docker/docker/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"go.uber.org/zap"
)

type Opt func(*Client) error

func WithAPIClient(apiClient client.APIClient) Opt {
	return func(c *Client) error {
		c.APIClient = apiClient
		return nil
	}
}

func WithVersionNegotiation(c *Client) error {
	c.dockerOpts = append(c.dockerOpts, client.WithAPIVersionNegotiation())
	return nil
}

func WithAuthProvider(authProvider provider.AuthProvider) Opt {
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

func WithImageProvider(imageProvider provider.ImageProvider) Opt {
	return func(c *Client) error {
		c.imageProvider = imageProvider
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
