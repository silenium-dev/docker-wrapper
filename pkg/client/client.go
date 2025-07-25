package client

import (
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"slices"
)

type AuthProvider interface {
	AuthConfig(named reference.Named) registry.AuthConfig
}

type Client struct {
	*client.Client
	dockerOpts   []client.Opt
	authProvider AuthProvider
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

func WithAuthProvider(authProvider AuthProvider) Opt {
	return func(c *Client) error {
		c.authProvider = authProvider
		return nil
	}
}

func WithDockerOpts(opts ...client.Opt) Opt {
	return func(c *Client) error {
		c.dockerOpts = slices.Concat(c.dockerOpts, opts)
		return nil
	}
}

func FromEnv(c *Client) error {
	c.dockerOpts = append(c.dockerOpts, client.FromEnv)
	return nil
}
