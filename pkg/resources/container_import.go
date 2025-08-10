package resources

import (
	"context"
)

func (c *Client) ImportContainer(ctx context.Context, id string) (*Container, error) {
	c.containersMutex.Lock()
	defer c.containersMutex.Unlock()

	inspect, err := c.wrapper.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	var cont *Container
	var ok bool
	if cont, ok = c.containers[inspect.ID]; !ok {
		cont = &Container{
			client: c,
			id:     inspect.ID,
			name:   inspect.Name,
		}
	}

	networks := make(map[string]*Network)
	for _, netConfig := range inspect.NetworkSettings.Networks {
		if netConfig.NetworkID == "" {
			continue // Default network
		}
		net, err := c.ImportNetwork(ctx, netConfig.NetworkID)
		if err != nil {
			return nil, err
		}
		networks[net.Id()] = net
	}

	volumes := make(map[string]*Volume)
	for _, volConfig := range inspect.Mounts {
		if volConfig.Type != "volume" {
			continue
		}
		vol, err := c.ImportVolume(ctx, volConfig.Name)
		if err != nil {
			return nil, err
		}
		volumes[vol.Name()] = vol
	}

	cont.networks = networks
	cont.volumes = volumes

	c.containers[inspect.ID] = cont
	return cont, nil
}
