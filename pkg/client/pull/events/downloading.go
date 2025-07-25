package events

import (
	"fmt"
)

type Downloading struct {
	ProgressBase
}

func (d *Downloading) String() string {
	return fmt.Sprintf("[%s] downloading %s", d.LayerId(), d.Progress())
}

const DownloadingStatus = "Downloading"
