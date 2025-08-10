package resources

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/silenium-dev/docker-wrapper/pkg/errors"
)

const volumeLabel = "dev.silenium.docker-wrapper.volume-access"

func (v *Volume) ensureAccessContainer(ctx context.Context, labels ResourceLabels) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.accessContainer != nil {
		return nil
	}

	name := fmt.Sprintf("%s-access", v.name)
	cont, err := v.client.ImportContainer(ctx, name)
	if errors.IsNotFound(err, errors.ResourceTypeContainer) {
		trimmed := name
		if labels != nil {
			trimmed = labels.TrimName(name)
		}
		cont, err = v.client.CreateContainer(
			ctx, trimmed, labels, ContainerSpec{
				Image: v.client.imageProvider.GetVolumeAccessImage(),
				Labels: map[string]string{
					volumeLabel: v.name,
				},
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: v.name,
						Target: "/mnt",
					},
				},
				Entrypoint: []string{"/bin/sh", "-c", "trap : TERM INT; sleep infinity & wait"},
			},
		)
	}
	if err != nil {
		return err
	}
	v.accessContainer = cont
	return nil
}

func (v *Volume) removeAccessContainer(ctx context.Context) error {
	if v.accessContainer == nil {
		return nil
	}
	err := v.accessContainer.Remove(ctx, container.RemoveOptions{Force: true})
	if err == nil {
		v.accessContainer = nil
	}
	return err
}
