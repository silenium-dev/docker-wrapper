package auth

import (
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"maps"
)

type OverridingAuthProvider struct {
	source    Provider
	overrides map[string]registry.AuthConfig
	config    *ProviderConfig
}

func NewOverridingProvider(
	source Provider, overrides map[string]registry.AuthConfig, opts ...Opt,
) *OverridingAuthProvider {
	config := renderConfig(opts)
	return &OverridingAuthProvider{source, overrides, config}
}

func (o *OverridingAuthProvider) WithOverride(domain string, ac registry.AuthConfig) *OverridingAuthProvider {
	overrides := maps.Clone(o.overrides)
	overrides[domain] = ac
	return NewOverridingProvider(o.source, overrides)
}

func (o *OverridingAuthProvider) AuthConfigs() map[string]registry.AuthConfig {
	authConfigs := o.source.AuthConfigs()
	maps.Copy(authConfigs, o.overrides)
	return authConfigs
}

func (o *OverridingAuthProvider) AuthConfig(ref reference.Named) registry.AuthConfig {
	if ac, ok := o.overrides[reference.Domain(ref)]; ok {
		o.config.Logger.Debugf("using override for %s", reference.Domain(ref))
		return ac
	}
	o.config.Logger.Debugf("no override present for %s", reference.Domain(ref))
	return o.source.AuthConfig(ref)
}

func (o *OverridingAuthProvider) Resolve(resource authn.Resource) (authn.Authenticator, error) {
	return &SimpleAuthenticator{AuthConfig: o.AuthConfigs()[resource.RegistryStr()]}, nil
}
