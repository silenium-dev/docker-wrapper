package events

import (
	"fmt"
	"github.com/opencontainers/go-digest"
)

type Digest struct {
	Digest digest.Digest
}

func (d *Digest) String() string {
	return fmt.Sprintf("Digest: %s", d.Digest.String())
}
