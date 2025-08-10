package resources

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

func (c *Client) ImportNetwork(ctx context.Context, id string) (*Network, error) {
	c.networksMutex.Lock()
	defer c.networksMutex.Unlock()

	inspect, err := c.wrapper.NetworkInspect(ctx, id, network.InspectOptions{Verbose: true})
	if err != nil {
		return nil, err
	}

	if net, ok := c.networks[inspect.ID]; ok {
		return net, nil
	}
	net := &Network{
		client: c,
		id:     inspect.ID,
		name:   inspect.Name,
	}
	c.networks[inspect.ID] = net
	return net, nil
}
