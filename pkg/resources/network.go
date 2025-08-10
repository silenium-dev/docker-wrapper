package resources

import (
	"context"
	"net"
	"sync"

	"github.com/silenium-dev/docker-wrapper/pkg/api"
)

type Network struct {
	client  *Client
	id      string
	name    string
	mutex   sync.RWMutex
	removed bool
}

func (n *Network) Id() string {
	return n.id
}

func (n *Network) Name() string {
	return n.name
}

func (n *Network) Client() api.ClientWrapper {
	return n.client.wrapper
}

func (n *Network) HostIP(ctx context.Context) (net.IP, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if err := n.ensureNotRemoved(); err != nil {
		return nil, err
	}

	return n.client.wrapper.SystemHostIPFromContainers(ctx, &n.id)
}

func (n *Network) ensureNotRemoved() error {
	if n.removed {
		return ErrResourceRemoved
	}
	return nil
}
