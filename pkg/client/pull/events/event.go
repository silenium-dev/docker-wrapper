package events

import (
	"fmt"
	"github.com/opencontainers/go-digest"
	"github.com/silenium-dev/docker-wrapper/pkg/client/pull/base"
	"strings"
	"time"
)

type PullEvent interface {
	String() string
}

type LayerEvent interface {
	LayerId() string
}

type FinalEvent interface {
	IsFinal() bool
	Message() string
}

type LayerBase struct {
	id string
}

func (l *LayerBase) LayerId() string {
	return l.id
}

type ProgressBase struct {
	LayerBase
	progress Progress
}

func (p *ProgressBase) Progress() Progress {
	return p.progress
}

func Parse(event base.PullProgressEvent) (PullEvent, error) {
	layer := LayerBase{id: event.ID}

	progress := Progress{
		Total:   event.ProgressDetail.Total,
		Current: event.ProgressDetail.Current,
		Hide:    event.ProgressDetail.HideCounts,
	}
	progressBase := ProgressBase{layer, progress}

	switch event.Status {
	case AlreadyExistsStatus:
		return &AlreadyExists{layer}, nil
	case PullingFSLayerStatus:
		return &PullingFSLayer{layer}, nil
	case WaitingStatus:
		return &Waiting{layer}, nil
	case DownloadingStatus:
		return &Downloading{progressBase}, nil
	case VerifyingChecksumStatus:
		return &VerifyingChecksum{layer}, nil
	case DownloadCompleteStatus:
		return &DownloadComplete{layer}, nil
	case ExtractingStatus:
		var duration time.Duration
		var err error
		if event.ProgressDetail.Units != "" {
			duration, err = time.ParseDuration(
				fmt.Sprintf("%d%s", event.ProgressDetail.Current, event.ProgressDetail.Units),
			)
			if err != nil {
				return nil, err
			}
		}
		return &Extracting{progressBase, duration}, nil
	case PullCompleteStatus:
		return &PullComplete{layer}, nil
	default:
		if strings.HasPrefix(event.Status, "Pulling from") {
			return &PullStarted{}, nil
		}
		if strings.HasPrefix(event.Status, "Digest:") {
			hash := digest.FromString(strings.TrimPrefix(event.Status, "Digest: "))
			return &Digest{hash}, nil
		}
		if strings.HasPrefix(event.Status, "Status:") {
			status := strings.TrimPrefix(event.Status, "Status: ")
			final := Final{status}
			if strings.HasPrefix(status, "Downloaded newer image") {
				return &DownloadedNewerImage{final}, nil
			}
			if strings.HasPrefix(status, "Image is up to date") {
				return &UpToDate{final}, nil
			}
			return &final, nil
		}
	}
	return nil, fmt.Errorf("unknown pull status: %s", event.Status)
}
