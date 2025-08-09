package provider

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/cpuguy83/dockercfg"
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
)

type DefaultAuthProvider struct {
	authConfigs map[string]registry.AuthConfig
	config      *AuthProviderConfig
}

func (d *DefaultAuthProvider) Resolve(resource authn.Resource) (authn.Authenticator, error) {
	return &SimpleAuthenticator{AuthConfig: d.AuthConfigs()[resource.RegistryStr()]}, nil
}

func (d *DefaultAuthProvider) AuthConfigs() map[string]registry.AuthConfig {
	return d.authConfigs
}

func (d *DefaultAuthProvider) AuthConfig(ref reference.Named) registry.AuthConfig {
	domain := reference.Domain(ref)
	if ac, ok := d.authConfigs[domain]; ok {
		d.config.Logger.Debugf("using auth config for %s", domain)
		return ac
	}
	d.config.Logger.Debugf("no auth config for %s", domain)
	return registry.AuthConfig{}
}

func NewDefaultAuthProvider(opts ...AuthOpt) (*DefaultAuthProvider, error) {
	config := renderConfig(opts)

	authConfigs := map[string]registry.AuthConfig{}

	cfg, err := dockercfg.LoadDefaultConfig()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		config.Logger.Errorf("failed to load docker config: %v", err)
		return nil, err
	} else if errors.Is(err, os.ErrNotExist) {
		return &DefaultAuthProvider{authConfigs, config}, nil
	}

	for k, v := range cfg.AuthConfigs {
		ac := registry.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Email:         v.Email,
			Auth:          v.Auth,
			IdentityToken: v.IdentityToken,
			RegistryToken: v.RegistryToken,
			ServerAddress: v.ServerAddress,
		}
		if ac.Username == "" && ac.Password == "" {
			err := getCredentials(k, &cfg, &ac)
			if err != nil {
				config.Logger.Errorf("failed to get credentials for registry %s: %v", k, err)
				continue
			}
		}
		if ac.Auth == "" {
			ac.Auth = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ac.Username, ac.Password)))
		}
		authConfigs[k] = ac
	}

	for k := range cfg.CredentialHelpers {
		ac := authConfigs[k]
		err := getCredentials(k, &cfg, &ac)
		if err != nil {
			config.Logger.Errorf("failed to get credentials for registry %s: %v", k, err)
		}
		authConfigs[k] = ac
	}
	return &DefaultAuthProvider{authConfigs, config}, nil
}

func getCredentials(registryHost string, cfg *dockercfg.Config, ac *registry.AuthConfig) error {
	u, p, err := cfg.GetRegistryCredentials(registryHost)
	if err != nil {
		return err
	}
	if u == "" {
		ac.IdentityToken = p
	} else {
		ac.Username = u
		ac.Password = p
	}

	return nil
}
