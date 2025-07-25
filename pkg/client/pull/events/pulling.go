package events

type PullStarted struct {
}

func (p *PullStarted) String() string {
	return "Pulling image"
}
