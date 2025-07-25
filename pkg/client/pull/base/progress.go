package base

// PullProgressEvent represents events returned from image pulling
type PullProgressEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current    int    `json:"current"`
		Total      int    `json:"total"`
		HideCounts bool   `json:"hidecounts,omitempty"`
		Units      string `json:"units,omitempty"`
	} `json:"progressDetail"`
}
