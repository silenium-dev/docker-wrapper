package state

import (
	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
)

type Pull interface {
	Ref() reference.Named
	Digest() *digest.Digest // can be nil when the pull is not yet complete
	Layers() []Layer
	Layer(id string) Layer
}

type Layer interface {
	Id() string
	Status() string
}

type pullBase struct {
	ref    reference.Named
	layers []Layer
}

func (p *pullBase) Ref() reference.Named {
	return p.ref
}

func (p *pullBase) Digest() *digest.Digest {
	return nil
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
