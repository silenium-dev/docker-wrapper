package events

import "fmt"

type DownloadComplete struct {
	LayerBase
}

func (d *DownloadComplete) String() string {
	return fmt.Sprintf("[%s] download complete", d.LayerId())
}

const DownloadCompleteStatus = "Download complete"
