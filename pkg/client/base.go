package client

import (
	"github.com/silenium-dev/docker-wrapper/pkg/client/provider"
	"go.uber.org/zap"
)

func (c *Client) Close() error {
	return c.APIClient.Close()
}

func (c *Client) AuthProvider() provider.AuthProvider {
	return c.authProvider
}

func (c *Client) Logger() *zap.SugaredLogger {
	return c.logger
}
