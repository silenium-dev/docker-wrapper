package image

type defaultImageProvider struct {
}

func (d *defaultImageProvider) GetDnsUtilImage() string {
	return "registry.k8s.io/e2e-test-images/agnhost:2.39"
}

func DefaultProvider() Provider {
	return &defaultImageProvider{}
}
