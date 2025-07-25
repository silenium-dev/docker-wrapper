package events

import "fmt"

type DownloadedNewerImage struct {
	Final
}

func (d *DownloadedNewerImage) String() string {
	return "Downloaded newer image"
}

type UpToDate struct {
	Final
}

func (u *UpToDate) String() string {
	return fmt.Sprintf("Image is up to date")
}

type Final struct {
	message string
}

func (f *Final) String() string {
	return fmt.Sprintf("Status: %s", f.Message())
}

func (f *Final) IsFinal() bool {
	return true
}
func (f *Final) Message() string {
	return f.message
}
