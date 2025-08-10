package resources

import (
	"strings"

	"github.com/docker/docker/api/types/filters"
)

type ResourceLabels interface {
	ToMap() map[string]string
	ToFilter() filters.Args
	// FullName adds an eventual prefix and/or suffix to the name
	FullName(name string) string
	// TrimName returns the name without any prefix or suffix
	TrimName(name string) string
}

type importedLabels struct {
	labels map[string]string
}

func (i *importedLabels) ToMap() map[string]string {
	return i.labels
}

func (i *importedLabels) ToFilter() filters.Args {
	result := strings.Builder{}
	for k, v := range i.labels {
		if result.Len() > 0 {
			result.WriteString(",")
		}
		result.WriteString(k)
		result.WriteString("=")
		result.WriteString(v)
	}
	return filters.NewArgs(filters.Arg("label", result.String()))
}

func (i *importedLabels) FullName(name string) string {
	return name
}

func (i *importedLabels) TrimName(name string) string {
	return name
}
