package events

import "fmt"

type PullingFSLayer struct {
	LayerBase
}

func (p *PullingFSLayer) String() string {
	return fmt.Sprintf("[%s] pulling fs layer", p.id)
}

const PullingFSLayerStatus = "Pulling fs layer"
