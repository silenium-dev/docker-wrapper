package state

import (
	"github.com/distribution/reference"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
)

type Pull interface {
	Ref() reference.Named
	Layers() []Layer
	Layer(id string) Layer
	Next(event events.PullEvent) (Pull, error)
	Status() string
}

type Layer interface {
	Id() string
	Status() string
	Next(event events.LayerEvent) (Layer, error)
}

type pullBase struct {
	ref    reference.Named
	layers []Layer
}

func (p *pullBase) Ref() reference.Named {
	return p.ref
}

func (p *pullBase) Layers() []Layer {
	return p.layers
}

func (p *pullBase) Layer(id string) Layer {
	for _, l := range p.layers {
		if l.Id() == id {
			return l
		}
	}
	return nil
}

type layerBase struct {
	id string
}

func (l *layerBase) Id() string {
	return l.id
}
