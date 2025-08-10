package errors

import (
	"fmt"
	"strings"

	"github.com/containerd/errdefs"
)

type ResourceType string

const (
	ResourceTypeContainer = "container"
	ResourceTypeVolume    = "volume"
	ResourceTypeImage     = "image"
)

func IsNotFound(err error, resource ResourceType) bool {
	if err == nil {
		return false
	}
	if errdefs.IsNotFound(err) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), fmt.Sprintf("no such %s", resource))
}
