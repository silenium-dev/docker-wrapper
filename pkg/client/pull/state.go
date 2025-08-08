package pull

import (
	"context"

	"github.com/distribution/reference"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/state"
)

func StateFromStream(
	ctx context.Context, ref reference.Named, isPodman bool, ch chan events.PullEvent, manifest *v1.Manifest,
	dig v1.Hash,
) chan state.Pull {
	out := make(chan state.Pull)

	go processEvents(ctx, ref, isPodman, ch, manifest, dig, out)

	return out
}

func processEvents(
	ctx context.Context, ref reference.Named, isPodman bool,
	ch chan events.PullEvent, manifest *v1.Manifest,
	dig v1.Hash, out chan state.Pull,
) {
	defer close(out)
	var current state.Pull
	var err error
	for {
		select {
		case <-ctx.Done():
			goto done
		case event, ok := <-ch:
			if !ok {
				goto done
			}
			var next state.Pull
			if current == nil {
				next, err = state.NewPullState(ref, isPodman, manifest, dig, event)
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
done:
	_, pullComplete := current.(*state.PullComplete)
	_, pullErrored := current.(*state.PullErrored)
	if !pullComplete && !pullErrored {
		success := true
		pulledNewer := false
		for _, l := range current.Layers() {
			_, isComplete := l.(*state.LayerPullComplete)
			_, isAlreadyExists := l.(*state.LayerAlreadyExists)
			if isComplete {
				pulledNewer = true
			}
			if !isComplete && !isAlreadyExists {
				success = false
				break
			}
		}
		if success {
			out <- &state.PullComplete{
				PullBase:        current.Base(),
				ImageDigest:     digest.Digest(dig.String()),
				DownloadedNewer: pulledNewer,
			}
		}
	}
}
