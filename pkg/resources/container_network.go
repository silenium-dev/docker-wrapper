package resources

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

func (c *Container) Connect(ctx context.Context, n *Network, config *network.EndpointSettings) error {
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	return n.Connect(ctx, c, config)
}

func (c *Container) Disconnect(ctx context.Context, n *Network, force bool) error {
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	return n.Disconnect(ctx, c, force)
}
