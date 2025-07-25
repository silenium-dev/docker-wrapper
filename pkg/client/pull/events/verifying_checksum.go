package events

import "fmt"

type VerifyingChecksum struct {
	LayerBase
}

func (v *VerifyingChecksum) String() string {
	return fmt.Sprintf("[%s] verifying checksum", v.id)
}

const VerifyingChecksumStatus = "Verifying Checksum"
