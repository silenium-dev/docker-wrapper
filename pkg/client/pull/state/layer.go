package state

import (
	"fmt"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/events"
	"time"
)

func NewLayer(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.PullingFSLayer:
		return &LayerPullingFSLayer{layerBase{event.LayerId()}}, nil
	case *events.AlreadyExists:
		return &LayerAlreadyExists{layerBase{event.LayerId()}}, nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid initial event (%T)", event)
}

type LayerErrored struct {
	layerBase
	error string
}

func (l *LayerErrored) Status() string {
	return fmt.Sprintf("Error: %s", l.error)
}

func (l *LayerErrored) Next(events.LayerEvent) (Layer, error) {
	return nil, fmt.Errorf("layer errored: %s", l.error)
}

type LayerPullingFSLayer struct {
	layerBase
}

func (l *LayerPullingFSLayer) Status() string {
	return "Pulling fs layer"
}

func (l *LayerPullingFSLayer) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.Waiting:
		return &LayerWaiting{l.layerBase}, nil
	case *events.Downloading:
		return &LayerDownloading{l.layerBase, event.Progress()}, nil
	case *events.DownloadComplete:
		return &LayerDownloadComplete{l.layerBase}, nil
	case *events.AlreadyExists:
		return &LayerAlreadyDownloaded{l.layerBase}, nil
	case *events.LayerError:
		return &LayerErrored{l.layerBase, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid transition (pulling-fs-layer + %T)", event)
}

type LayerWaiting struct {
	layerBase
}

func (l *LayerWaiting) Status() string {
	return "Waiting"
}

func (l *LayerWaiting) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.AlreadyExists:
		return &LayerAlreadyDownloaded{l.layerBase}, nil
	case *events.Downloading:
		return &LayerDownloading{l.layerBase, event.Progress()}, nil
	case *events.DownloadComplete:
		return &LayerDownloadComplete{l.layerBase}, nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid transition (waiting + %T)", event)
}

type LayerDownloading struct {
	layerBase
	progress events.Progress
}

func (l *LayerDownloading) Status() string {
	return fmt.Sprintf("Downloading (%s)", l.progress.String())
}

func (l *LayerDownloading) Progress() events.Progress {
	return l.progress
}

func (l *LayerDownloading) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.Downloading:
		return &LayerDownloading{l.layerBase, event.Progress()}, nil
	case *events.DownloadComplete:
		return &LayerDownloadComplete{l.layerBase}, nil
	case *events.VerifyingChecksum:
		return &LayerVerifyingChecksum{l.layerBase}, nil
	case *events.Extracting:
		return parseLayerExtracting(l.layerBase, event), nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid transition (downloading + %T)", event)
}

type LayerVerifyingChecksum struct {
	layerBase
}

func (l *LayerVerifyingChecksum) Status() string {
	return "Verifying Checksum"
}

func (l *LayerVerifyingChecksum) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.DownloadComplete:
		return &LayerDownloadComplete{l.layerBase}, nil
	case *events.Extracting:
		return parseLayerExtracting(l.layerBase, event), nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid transition (verifying-checksum + %T)", event)
}

type LayerDownloadComplete struct {
	layerBase
}

func (l *LayerDownloadComplete) Status() string {
	return "Download complete"
}

func (l *LayerDownloadComplete) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.Extracting:
		return parseLayerExtracting(l.layerBase, event), nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid transition (download-complete + %T)", event)
}

type LayerAlreadyDownloaded struct {
	layerBase
}

func (l *LayerAlreadyDownloaded) Status() string {
	return "Already downloaded"
}

func (l *LayerAlreadyDownloaded) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.Extracting:
		return parseLayerExtracting(l.layerBase, event), nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	case *events.PullComplete:
		return &LayerPullComplete{l.layerBase}, nil
	case *events.AlreadyExists:
		return &LayerAlreadyExists{l.layerBase}, nil
	}
	return nil, fmt.Errorf("invalid transition (already-downloaded + %T)", event)
}

type LayerExtracting struct {
	layerBase
	duration *time.Duration
	progress *events.Progress
}

func (l *LayerExtracting) Status() string {
	switch {
	case l.duration != nil:
		return fmt.Sprintf("Extracting (%s)", l.duration.String())
	case l.progress != nil:
		return fmt.Sprintf("Extracting (%s)", l.progress.String())
	default:
		return "Extracting"
	}
}

func (l *LayerExtracting) Progress() *events.Progress {
	return l.progress
}

func (l *LayerExtracting) Duration() *time.Duration {
	return l.duration
}

func (l *LayerExtracting) Next(event events.LayerEvent) (Layer, error) {
	switch event := event.(type) {
	case *events.Extracting:
		return parseLayerExtracting(l.layerBase, event), nil
	case *events.PullComplete:
		return &LayerPullComplete{l.layerBase}, nil
	case *events.LayerError:
		return &LayerErrored{layerBase{event.LayerId()}, event.Error}, nil
	}
	return nil, fmt.Errorf("invalid transition (extracting + %T)", event)
}

func parseLayerExtracting(base layerBase, event *events.Extracting) Layer {
	progress := event.Progress()
	duration := event.Duration
	if progress.Total == 0 {
		return &LayerExtracting{base, &duration, nil}
	} else {
		return &LayerExtracting{base, nil, &progress}
	}
}

// Final states

type LayerAlreadyExists struct {
	layerBase
}

func (l *LayerAlreadyExists) Status() string {
	return "Already exists"
}

func (l *LayerAlreadyExists) Next(event events.LayerEvent) (Layer, error) {
	return nil, fmt.Errorf("already completed, tried %T on layer-already-exists", event)
}

type LayerPullComplete struct {
	layerBase
}

func (l *LayerPullComplete) Status() string {
	return "Pull complete"
}

func (l *LayerPullComplete) Next(event events.LayerEvent) (Layer, error) {
	return nil, fmt.Errorf("already completed, tried %T on layer-pull-complete", event)
}
