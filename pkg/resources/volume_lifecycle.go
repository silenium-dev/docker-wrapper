package resources

import "context"

func (v *Volume) Remove(ctx context.Context) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if err := v.ensureNotRemoved(); err != nil {
		return err
	}

	if err := v.removeAccessContainer(ctx); err != nil {
		return err
	}

	v.client.volumesMutex.Lock()
	defer v.client.volumesMutex.Unlock()
	err := v.client.wrapper.VolumeRemove(ctx, v.name, true)
	if err == nil {
		delete(v.client.volumes, v.name)
		v.removed = true
	}
	return err
}
