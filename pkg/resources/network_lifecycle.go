package resources

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

func (n *Network) Remove(ctx context.Context) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if err := n.ensureNotRemoved(); err != nil {
		return err
	}

	n.client.networksMutex.Lock()
	defer n.client.networksMutex.Unlock()
	err := n.client.wrapper.NetworkRemove(ctx, n.id)
	if err == nil {
		delete(n.client.networks, n.id)
		n.removed = true
	}
	return err
}

func (n *Network) Connect(ctx context.Context, container *Container, config *network.EndpointSettings) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if err := n.ensureNotRemoved(); err != nil {
		return err
	}

	err := n.client.wrapper.NetworkConnect(ctx, n.id, container.id, config)
	if err == nil {
		container.networks[n.id] = n
	}
	return err
}

func (n *Network) Disconnect(ctx context.Context, container *Container, force bool) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if err := n.ensureNotRemoved(); err != nil {
		return err
	}

	err := n.client.wrapper.NetworkDisconnect(ctx, n.id, container.id, force)
	if err == nil {
		delete(container.networks, n.id)
	}
	return err
}
