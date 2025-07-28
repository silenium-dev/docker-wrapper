package state

import (
	"github.com/distribution/reference"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
)

type Pull interface {
	Ref() reference.Named
	Manifest() *v1.Manifest
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
	ref      reference.Named
	layers   map[string]Layer
	manifest *v1.Manifest
}

func (p *pullBase) Ref() reference.Named {
	return p.ref
}

func (p *pullBase) Manifest() *v1.Manifest {
	return p.manifest
}

func (p *pullBase) Layers() []Layer {
	layers := make([]Layer, 0, len(p.layers))
	for _, l := range p.manifest.Layers {
		for i := 1; i <= len(l.Digest.Hex); i++ {
			layer, ok := p.layers[l.Digest.Hex[:i]]
			if ok {
				layers = append(layers, layer)
				break
			}
		}
	}
	return layers
}

func (p *pullBase) Layer(id string) Layer {
	return p.layers[id]
}

type layerBase struct {
	id string
}

func (l *layerBase) Id() string {
	return l.id
}
