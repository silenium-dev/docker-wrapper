package events

type LayerError struct {
	LayerBase
	PullError
}

type PullError struct {
	Error string
}

func (e *PullError) String() string {
	return e.Error
}
