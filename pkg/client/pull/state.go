package pull

import (
	"context"
	"github.com/distribution/reference"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
)

func StateFromStream(ctx context.Context, ref reference.Named, ch chan events.PullEvent) chan state.Pull {
	out := make(chan state.Pull)

	go processEvents(ctx, ref, ch, out)

	return out
}

func processEvents(ctx context.Context, ref reference.Named, ch chan events.PullEvent, out chan state.Pull) {
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
				next, err = state.NewPullState(ref, event)
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
