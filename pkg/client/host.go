package client

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"k8s.io/apimachinery/pkg/util/rand"
)

func (c *Client) HostIPFromContainers(ctx context.Context, netId *string) (net.IP, error) {
	c.hostFromContainerMutex.RLock()
	if c.hostFromContainerAddr != nil { // Fast path
		defer c.hostFromContainerMutex.RUnlock()
		return c.hostFromContainerAddr, nil
	}
	c.hostFromContainerMutex.RUnlock()
	c.hostFromContainerMutex.Lock()
	defer c.hostFromContainerMutex.Unlock()
	if c.hostFromContainerAddr != nil { // Recheck
		return c.hostFromContainerAddr, nil
	}

	isPodman, err := c.IsPodman(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if Podman: %w", err)
	}

	endpoints := map[string]*network.EndpointSettings{}
	if netId != nil {
		endpoints[*netId] = &network.EndpointSettings{}
	}

	cont, err := c.ContainerCreate(
		ctx,
		&container.Config{
			Image:      c.imageProvider.GetDnsUtilImage(),
			Entrypoint: []string{"dig", "+short", "A", "host.docker.internal"},
		},
		&container.HostConfig{},
		&network.NetworkingConfig{EndpointsConfig: endpoints},
		nil,
		rand.String(16),
	)
	if err != nil {
		return nil, err
	}
	defer c.ContainerRemove(context.Background(), cont.ID, container.RemoveOptions{Force: true})

	err = c.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return nil, err
	}

	inspect, err := c.ContainerInspect(ctx, cont.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", cont.ID, err)
	}

	multiplex, err := c.StreamLogs(ctx, cont.ID, true)
	if err != nil {
		return nil, err
	}
	ipAddrStr, ok := <-multiplex.Stdout()
	if !ok {
		return nil, fmt.Errorf("no output from container %s", cont.ID)
	}
	if !isPodman {
		ipAddrStr = []byte(inspect.NetworkSettings.Gateway)
	}

	ipAddr := net.ParseIP(strings.TrimSpace(string(ipAddrStr)))
	if ipAddr == nil {
		return nil, fmt.Errorf("failed to parse IP address from: %s", ipAddrStr)
	}

	c.hostFromContainerAddr = ipAddr
	return c.hostFromContainerAddr, nil
}
