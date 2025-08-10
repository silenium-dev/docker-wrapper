package resources

import (
	"context"
	"errors"
	"sync"

	"github.com/docker/docker/api/types/network"
	"github.com/silenium-dev/docker-wrapper/pkg/api"
	"github.com/silenium-dev/docker-wrapper/pkg/client/stream"
)

type Container struct {
	client            *Client
	id                string
	name              string
	mutex             sync.RWMutex
	restartInProgress bool
	removed           bool
	networks          map[string]*Network
	networkMutex      sync.RWMutex
	volumes           map[string]*Volume
}

func (c *Container) Id() string {
	return c.id
}

func (c *Container) Name() string {
	return c.name
}

func (c *Container) Client() api.ClientWrapper {
	return c.client.wrapper
}

func (c *Container) IsRestartInProgress() bool {
	return c.restartInProgress
}

func (c *Container) IsRunning(ctx context.Context) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return false, err
	}

	inspect, err := c.client.wrapper.ContainerInspect(ctx, c.id)
	if err != nil {
		return false, err
	}
	return inspect.State.Running, nil
}

func (c *Container) Logs(ctx context.Context, follow bool) (*stream.MultiplexedStream, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return nil, err
	}

	return c.client.wrapper.StreamLogs(ctx, c.id, follow)
}

func (c *Container) ensureNotRemoved() error {
	if c.removed {
		return ErrResourceRemoved
	}
	return nil
}

var ErrNotAttachedToNetwork = errors.New("container is not attached to network")

func (c *Container) GetEndpoint(ctx context.Context, network *Network) (*network.EndpointSettings, error) {
	inspect, err := c.client.wrapper.ContainerInspect(ctx, c.id)
	if err != nil {
		return nil, err
	}
	endpoint, ok := inspect.NetworkSettings.Networks[network.name]
	if !ok {
		return nil, ErrNotAttachedToNetwork
	}
	return endpoint, nil
}
