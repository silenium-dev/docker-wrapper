package events

import "fmt"

type Waiting struct {
	LayerBase
}

func (w *Waiting) String() string {
	return fmt.Sprintf("[%s] waiting", w.id)
}

const WaitingStatus = "Waiting"
