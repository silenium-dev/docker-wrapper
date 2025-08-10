package resources

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

func (c *Container) Start(ctx context.Context, options container.StartOptions) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return err
	}

	return c.client.wrapper.ContainerStart(ctx, c.id, options)
}

func (c *Container) Stop(ctx context.Context, options container.StopOptions) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return err
	}

	return c.client.wrapper.ContainerStop(ctx, c.id, options)
}

func (c *Container) Kill(ctx context.Context, signal string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return err
	}

	return c.client.wrapper.ContainerKill(ctx, c.id, signal)
}

func (c *Container) Restart(ctx context.Context, options container.StopOptions) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return err
	}

	c.restartInProgress = true
	err = c.client.wrapper.ContainerRestart(ctx, c.id, options)
	c.restartInProgress = false
	return err
}

func (c *Container) Remove(ctx context.Context, options container.RemoveOptions) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.ensureNotRemoved()
	if err != nil {
		return err
	}

	c.client.containersMutex.Lock()
	defer c.client.containersMutex.Unlock()
	err = c.client.wrapper.ContainerRemove(ctx, c.id, options)
	if err == nil {
		delete(c.client.containers, c.id)
		c.removed = true
	}
	return err
}
