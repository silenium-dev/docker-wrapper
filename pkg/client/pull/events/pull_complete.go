package events

import "fmt"

type PullComplete struct {
	LayerBase
}

func (p *PullComplete) String() string {
	return fmt.Sprintf("[%s] pull complete", p.LayerId())
}

const PullCompleteStatus = "Pull complete"
