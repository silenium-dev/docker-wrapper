package pull

import (
	"context"
	"github.com/distribution/reference"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
)

func StateFromStream(ctx context.Context, ref reference.Named, ch chan events.PullEvent, manifest *v1.Manifest, dig v1.Hash) chan state.Pull {
	out := make(chan state.Pull)

	go processEvents(ctx, ref, ch, manifest, dig, out)

	return out
}

func processEvents(ctx context.Context, ref reference.Named, ch chan events.PullEvent, manifest *v1.Manifest, dig v1.Hash, out chan state.Pull) {
	defer close(out)
	var current state.Pull
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			var next state.Pull
			if current == nil {
				next, err = state.NewPullState(ref, manifest, dig, event)
			} else {
				next, err = current.Next(event)
			}
			if err != nil {
				panic(err)
			}
			current = next
			out <- current
		}
	}
}
