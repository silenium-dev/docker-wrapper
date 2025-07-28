package state

import (
	"fmt"
	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
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

func NewPullState(ref reference.Named, event events.PullEvent) (Pull, error) {
	switch event.(type) {
	case *events.PullStarted:
		return &PullInProgress{
			pullBase: pullBase{
				ref: ref,
			},
		}, nil
	}
	return nil, fmt.Errorf("invalid initial event (%T)", event)
}

func (p *PullInProgress) Next(event events.PullEvent) (Pull, error) {
	layers := p.layers
	if le, ok := event.(events.LayerEvent); ok {
		layers = make([]Layer, len(p.layers))
		copy(layers, p.layers)
		found := false
		for i, l := range p.layers {
			if l.Id() == le.LayerId() {
				found = true
				newL, err := l.Next(le)
				if err != nil {
					return nil, err
				}
				layers[i] = newL
				break
			}
		}
		if !found {
			layer, err := NewLayer(le)
			if err != nil {
				return nil, err
			}
			layers = append(layers, layer)
		}
	}

	var result Pull
	switch event := event.(type) {
	case events.LayerEvent:
		result = &PullInProgress{
			pullBase: pullBase{
				ref:    p.ref,
				layers: layers,
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
