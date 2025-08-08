package state

import (
	"fmt"
	"maps"

	"github.com/distribution/reference"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
)

type PullInProgress struct {
	PullBase
	digest *digest.Digest
}

func (p *PullInProgress) Status() string {
	if p.digest == nil {
		return "Pulling"
	}
	return "Finishing"
}

func NewPullState(ref reference.Named, isPodman bool, manifest *v1.Manifest, dig v1.Hash, event events.PullEvent) (
	Pull, error,
) {
	base := PullBase{
		ref:      ref,
		manifest: manifest,
		digest:   dig,
		layers:   make(map[string]Layer),
		isPodman: isPodman,
	}
	switch event := event.(type) {
	case *events.PullStarted:
		return &PullInProgress{
			PullBase: base,
		}, nil
	case events.LayerEvent:
		var err error
		base.layers[event.LayerId()], err = NewLayer(isPodman, event)
		if err != nil {
			return nil, err
		}
		return &PullInProgress{
			PullBase: base,
		}, nil
	}
	return nil, fmt.Errorf("invalid initial event (%T)", event)
}

func (p *PullInProgress) Next(event events.PullEvent) (Pull, error) {
	layers := p.layers
	if le, ok := event.(events.LayerEvent); ok {
		layers = maps.Clone(p.layers)
		layer, found := layers[le.LayerId()]
		if found {
			found = true
			newL, err := layer.Next(le)
			if err != nil {
				return nil, err
			}
			layers[layer.Id()] = newL
		} else {
			layer, err := NewLayer(p.isPodman, le)
			if err != nil {
				return nil, err
			}
			layers[layer.Id()] = layer
		}
	}

	var result Pull
	switch event := event.(type) {
	case events.LayerEvent:
		result = &PullInProgress{
			PullBase: PullBase{
				ref:      p.ref,
				layers:   layers,
				manifest: p.manifest,
				isPodman: p.isPodman,
			},
			digest: p.digest,
		}
	case *events.PullStarted:
		result = &PullInProgress{
			PullBase: p.PullBase,
			digest:   p.digest,
		}
	case *events.Digest:
		result = &PullInProgress{
			p.PullBase,
			&event.Digest,
		}
	case *events.PullError:
		result = &PullErrored{
			PullBase: p.PullBase,
			error:    event.Error,
		}
	case *events.DownloadedNewerImage:
		if p.digest == nil {
			return nil, fmt.Errorf("cannot complete pull: no digest event received")
		}
		result = &PullComplete{
			PullBase:        p.PullBase,
			ImageDigest:     *p.digest,
			DownloadedNewer: true,
		}
	case events.FinalEvent:
		if p.digest == nil {
			return nil, fmt.Errorf("cannot complete pull: no digest event received")
		}
		result = &PullComplete{
			PullBase:        p.PullBase,
			ImageDigest:     *p.digest,
			DownloadedNewer: false,
		}
	}

	return result, nil
}

type PullErrored struct {
	PullBase
	error string
}

func (p *PullErrored) Status() string {
	return fmt.Sprintf("Error: %s", p.error)
}

func (p *PullErrored) Next(events.PullEvent) (Pull, error) {
	return nil, fmt.Errorf("pull errored: %s", p.error)
}

type PullComplete struct {
	PullBase
	ImageDigest     digest.Digest
	DownloadedNewer bool
}

func (p *PullComplete) Status() string {
	return fmt.Sprintf("Complete (Digest: %s)", p.ImageDigest.String())
}

func (p *PullComplete) Next(event events.PullEvent) (Pull, error) {
	return nil, fmt.Errorf("pull already complete (event: %T)", event)
}

func (p *PullComplete) HasDownloadedNewer() bool {
	return p.DownloadedNewer
}
