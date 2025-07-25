package client

import (
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

type AuthProvider interface {
	AuthConfig(named reference.Named) registry.AuthConfig
}

type Client struct {
	*client.Client
	authProvider AuthProvider
}

func New(opts ...client.Opt) (*Client, error) {
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: cli,
	}, nil
}

func (c *Client) Close() error {
	return c.Client.Close()
}
