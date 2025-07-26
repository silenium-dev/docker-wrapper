package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/cpuguy83/dockercfg"
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"os"
)

type DefaultAuthProvider struct {
	authConfigs map[string]*registry.AuthConfig
	config      *ProviderConfig
}

func (d *DefaultAuthProvider) AuthConfig(ref reference.Named) registry.AuthConfig {
	domain := reference.Domain(ref)
	if ac, ok := d.authConfigs[domain]; ok {
		d.config.Logger.Debugf("using auth config for %s", domain)
		return *ac
	}
	d.config.Logger.Debugf("no auth config for %s", domain)
	return registry.AuthConfig{}
}

func NewDefaultProvider(opts ...Opt) (*DefaultAuthProvider, error) {
	config := renderConfig(opts)

	authConfigs := map[string]*registry.AuthConfig{}

	cfg, err := dockercfg.LoadDefaultConfig()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		config.Logger.Errorf("failed to load docker config: %v", err)
		return nil, err
	} else if errors.Is(err, os.ErrNotExist) {
		return &DefaultAuthProvider{authConfigs, config}, nil
	}

	for k, v := range cfg.AuthConfigs {
		ac := &registry.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Email:         v.Email,
			Auth:          v.Auth,
			IdentityToken: v.IdentityToken,
			RegistryToken: v.RegistryToken,
			ServerAddress: v.ServerAddress,
		}
		if ac.Username == "" && ac.Password == "" {
			err := getCredentials(k, &cfg, ac)
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
		err := getCredentials(k, &cfg, authConfigs[k])
		if err != nil {
			config.Logger.Errorf("failed to get credentials for registry %s: %v", k, err)
		}
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
