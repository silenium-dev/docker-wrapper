package events

import (
	"fmt"
	"time"
)

type Extracting struct {
	ProgressBase
	Duration time.Duration
}

func (d *Extracting) HasProgress() bool {
	return d.Progress().Total != 0
}

func (d *Extracting) String() string {
	if d.Progress().Total == 0 { // containerd snapshot image store
		return fmt.Sprintf("[%s] extracting %s", d.LayerId(), d.Duration.String())
	} else { // classic image store
		return fmt.Sprintf("[%s] extracting %s", d.LayerId(), d.Progress().String())
	}
}

const ExtractingStatus = "Extracting"
