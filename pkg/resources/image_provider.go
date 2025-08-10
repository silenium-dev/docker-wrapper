package resources

type ImageProvider interface {
	GetGrafanaImage() string
	GetLokiImage() string
	GetAlloyImage() string
	GetVolumeAccessImage() string
}

type defaultImageProvider struct {
}

func (d *defaultImageProvider) GetGrafanaImage() string {
	return "grafana/grafana:latest"
}

func (d *defaultImageProvider) GetLokiImage() string {
	return "grafana/loki:latest"
}

func (d *defaultImageProvider) GetAlloyImage() string {
	return "grafana/alloy:latest"
}

func (d *defaultImageProvider) GetVolumeAccessImage() string {
	return "alpine:latest"
}

func DefaultImageProvider() ImageProvider {
	return &defaultImageProvider{}
}
