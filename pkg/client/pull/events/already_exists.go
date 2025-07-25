package events

import "fmt"

type AlreadyExists struct {
	LayerBase
}

func (a *AlreadyExists) String() string {
	return fmt.Sprintf("[%s] already exists", a.id)
}

func (a *AlreadyExists) LayerId() string {
	return a.id
}

const AlreadyExistsStatus = "Already exists"
