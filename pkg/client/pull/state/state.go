package state

import (
	"github.com/distribution/reference"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
)

type Pull interface {
	Ref() reference.Named
	Manifest() *v1.Manifest
	Digest() digest.Digest
	Layers() []Layer
	Layer(id string) Layer
	Next(event events.PullEvent) (Pull, error)
	Status() string
	Base() PullBase
}

type Layer interface {
	Id() string
	Status() string
	Next(event events.LayerEvent) (Layer, error)
}

type PullBase struct {
	ref      reference.Named
	layers   map[string]Layer
	manifest *v1.Manifest
	digest   v1.Hash
	isPodman bool
}

func (p *PullBase) Base() PullBase {
	return *p
}

func (p *PullBase) Ref() reference.Named {
	return p.ref
}

func (p *PullBase) Manifest() *v1.Manifest {
	return p.manifest
}

func (p *PullBase) Digest() digest.Digest {
	return digest.Digest(p.digest.String())
}

func (p *PullBase) Layers() []Layer {
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

func (p *PullBase) Layer(id string) Layer {
	return p.layers[id]
}

type layerBase struct {
	id       string
	isPodman bool
}

func (l *layerBase) Id() string {
	return l.id
}
