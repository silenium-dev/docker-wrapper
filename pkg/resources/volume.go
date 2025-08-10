package resources

import (
	"sync"

	"github.com/silenium-dev/docker-wrapper/pkg/api"
)

type Volume struct {
	client          *Client
	name            string
	removed         bool
	mutex           sync.RWMutex
	labels          ResourceLabels
	accessContainer *Container
}

func (v *Volume) Name() string {
	return v.name
}

func (v *Volume) Client() api.ClientWrapper {
	return v.client.wrapper
}

func (v *Volume) ensureNotRemoved() error {
	if v.removed {
		return ErrResourceRemoved
	}
	return nil
}
