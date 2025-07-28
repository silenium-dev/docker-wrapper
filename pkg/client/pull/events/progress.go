package events

import (
	"fmt"
	"github.com/docker/go-units"
)

type Progress struct {
	Current int
	Total   int
	Hide    bool
}

func (p Progress) String() string {
	return fmt.Sprintf("%s/%s", p.HumanCurrent(), p.HumanTotal())
}

func (p Progress) HumanCurrent() string {
	return units.HumanSize(float64(p.Current))
}

func (p Progress) HumanTotal() string {
	return units.HumanSize(float64(p.Total))
}

func (p Progress) Number() float32 {
	if p.Total == 0 {
		return 0
	}
	return float32(p.Current) / float32(p.Total)
}
