package provider

type ImageProvider interface {
	// GetDnsUtilImage returns an OCI image having dig preinstalled (for example: "registry.k8s.io/e2e-test-images/agnhost:2.39")
	GetDnsUtilImage() string
}
