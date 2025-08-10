package resources

import "context"

func (c *Client) ImportVolume(ctx context.Context, name string) (*Volume, error) {
	c.volumesMutex.Lock()
	defer c.volumesMutex.Unlock()

	inspect, err := c.wrapper.VolumeInspect(ctx, name)
	if err != nil {
		return nil, err
	}

	if vol, ok := c.volumes[inspect.Name]; ok {
		return vol, nil
	}
	vol := &Volume{
		client: c,
		name:   inspect.Name,
		labels: &importedLabels{labels: inspect.Labels},
	}

	c.volumes[inspect.Name] = vol
	return vol, nil
}
