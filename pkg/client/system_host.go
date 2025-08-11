package client

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/silenium-dev/docker-wrapper/pkg/client/stream"
	"k8s.io/apimachinery/pkg/util/rand"
)

func (c *Client) SystemHostIPFromContainers(ctx context.Context, netId *string) (net.IP, error) {
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

	isPodman, err := c.SystemIsPodman(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if Podman: %w", err)
	}

	endpoints := map[string]*network.EndpointSettings{}
	if netId != nil {
		endpoints[*netId] = &network.EndpointSettings{}
	}

	command := []string{"sh", "-c", "sleep infinity"}
	if isPodman {
		command = []string{"dig", "+short", "A", "host.docker.internal"}
	}

	cont, err := c.ContainerCreate(
		ctx,
		&container.Config{
			Image:      c.imageProvider.GetDnsUtilImage(),
			Entrypoint: command,
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
	var ipAddrStr string
	if isPodman {
		multiplex, err := c.StreamLogs(ctx, cont.ID, true)
		if err != nil {
			return nil, err
		}
		ipAddrByteMsg, ok := <-multiplex.Messages()
		if !ok {
			return nil, fmt.Errorf("no output from container %s", cont.ID)
		}
		if ipAddrByteMsg.StreamType != stream.TypeStdout {
			return nil, fmt.Errorf("unexpected stream type %s from container %s", ipAddrByteMsg.StreamType.Name(), cont.ID)
		}
		ipAddrStr = strings.TrimSpace(string(ipAddrByteMsg.Content))
	} else if netId == nil {
		ipAddrStr = inspect.NetworkSettings.Gateway
	} else {
		endpoint, ok := inspect.NetworkSettings.Networks[*netId]
		if !ok {
			for _, v := range inspect.NetworkSettings.Networks {
				if v.NetworkID == *netId {
					endpoint = v
					ok = true
					break
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("network %s not found in container %s", *netId, cont.ID)
		}
		ipAddrStr = endpoint.Gateway
	}

	ipAddr := net.ParseIP(ipAddrStr)
	if ipAddr == nil {
		return nil, fmt.Errorf("failed to parse IP address from: %s", ipAddrStr)
	}

	c.hostFromContainerAddr = ipAddr
	return c.hostFromContainerAddr, nil
}
