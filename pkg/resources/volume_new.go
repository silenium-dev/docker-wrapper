package resources

import (
	"context"
	"maps"

	"github.com/docker/docker/api/types/volume"
)

func (c *Client) CreateVolume(ctx context.Context, labels ResourceLabels, name string) (*Volume, error) {
	allLabels := map[string]string{}
	fullName := name
	if labels != nil {
		maps.Copy(allLabels, labels.ToMap())
		fullName = labels.FullName(name)
	}

	c.volumesMutex.Lock()
	defer c.volumesMutex.Unlock()

	resp, err := c.wrapper.VolumeCreate(
		ctx, volume.CreateOptions{
			Name:   fullName,
			Labels: allLabels,
		},
	)
	if err != nil {
		return nil, err
	}

	v := &Volume{
		client: c,
		name:   resp.Name,
	}
	c.volumes[resp.Name] = v
	return v, nil
}
