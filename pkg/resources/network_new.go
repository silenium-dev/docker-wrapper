package resources

import (
	"context"
	"maps"

	"github.com/docker/docker/api/types/network"
)

func (c *Client) CreateNetwork(
	ctx context.Context, labels ResourceLabels, name string, options network.CreateOptions,
) (*Network, error) {
	fullName := name
	if labels != nil {
		fullName = labels.FullName(name)
		userLabels := options.Labels
		options.Labels = make(map[string]string, len(userLabels))
		maps.Copy(options.Labels, labels.ToMap())
		maps.Copy(options.Labels, userLabels)
	}
	c.networksMutex.Lock()
	defer c.networksMutex.Unlock()

	resp, err := c.wrapper.NetworkCreate(ctx, fullName, options)
	if err != nil {
		return nil, err
	}
	net := &Network{
		client: c,
		id:     resp.ID,
		name:   name,
	}
	c.networks[name] = net
	return net, nil
}
