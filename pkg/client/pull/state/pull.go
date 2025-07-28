package state

import (
	"fmt"
	"github.com/distribution/reference"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"maps"
)

type PullInProgress struct {
	pullBase
	digest *digest.Digest
}

func (p *PullInProgress) Status() string {
	if p.digest == nil {
		return "Pulling"
	}
	return "Finishing"
}

func NewPullState(ref reference.Named, manifest *v1.Manifest, event events.PullEvent) (Pull, error) {
	switch event.(type) {
	case *events.PullStarted:
		return &PullInProgress{
			pullBase: pullBase{
				ref:      ref,
				manifest: manifest,
				layers:   make(map[string]Layer),
			},
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
			layer, err := NewLayer(le)
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
			pullBase: pullBase{
				ref:      p.ref,
				layers:   layers,
				manifest: p.manifest,
			},
			digest: p.digest,
		}
	case *events.PullStarted:
		result = &PullInProgress{
			pullBase: p.pullBase,
			digest:   p.digest,
		}
	case *events.Digest:
		result = &PullInProgress{
			p.pullBase,
			&event.Digest,
		}
	case *events.PullError:
		result = &PullErrored{
			pullBase: p.pullBase,
			error:    event.Error,
		}
	case *events.DownloadedNewerImage:
		if p.digest == nil {
			return nil, fmt.Errorf("cannot complete pull: no digest event received")
		}
		result = &PullComplete{
			pullBase:        p.pullBase,
			digest:          *p.digest,
			downloadedNewer: true,
		}
	case events.FinalEvent:
		if p.digest == nil {
			return nil, fmt.Errorf("cannot complete pull: no digest event received")
		}
		result = &PullComplete{
			pullBase:        p.pullBase,
			digest:          *p.digest,
			downloadedNewer: false,
		}
	}

	return result, nil
}

type PullErrored struct {
	pullBase
	error string
}

func (p *PullErrored) Status() string {
	return fmt.Sprintf("Error: %s", p.error)
}

func (p *PullErrored) Next(events.PullEvent) (Pull, error) {
	return nil, fmt.Errorf("pull errored: %s", p.error)
}

type PullComplete struct {
	pullBase
	digest          digest.Digest
	downloadedNewer bool
}

func (p *PullComplete) Status() string {
	return fmt.Sprintf("Complete (Digest: %s)", p.digest.String())
}

func (p *PullComplete) Next(event events.PullEvent) (Pull, error) {
	return nil, fmt.Errorf("pull already complete (event: %T)", event)
}

func (p *PullComplete) Digest() digest.Digest {
	return p.digest
}

func (p *PullComplete) HasDownloadedNewer() bool {
	return p.downloadedNewer
}
