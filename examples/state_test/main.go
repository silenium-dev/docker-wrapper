package main

import (
	"context"
	"fmt"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/image"
	"github.com/silenium-dev/docker-wrapper/pkg/client"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, _ := client.NewWithOpts(client.FromEnv, client.WithVersionNegotiation)
	ref, _ := reference.ParseNormalizedNamed("localstack/localstack:3")
	id, manifest, eventChan, err := cli.ImagePullWithEvents(ctx, ref, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Pulling:", id.String(), manifest.MediaType)
	layerTracker := map[string][]events.PullEvent{}
	var tracker []events.PullEvent
	for event := range eventChan {
		if le, ok := event.(events.LayerEvent); ok {
			layerTracker[le.LayerId()] = append(layerTracker[le.LayerId()], event)
		} else {
			tracker = append(tracker, event)
		}
	}
	fmt.Println("\nPull Events:")
	for _, pullEvent := range tracker {
		fmt.Printf("  %s\n", pullEvent.String())
	}
	fmt.Println("\nLayer Events:")
	for _, layer := range manifest.Layers {
		fmt.Printf("Layer %s:\n", layer.Digest.String())
		layerEvents := layerTracker[layer.Digest.Hex[:12]]
		for _, event := range layerEvents {
			fmt.Printf("  %s\n", event.String())
		}
	}
}
