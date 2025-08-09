package provider

import (
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"go.uber.org/zap"
)

type AuthProvider interface {
	authn.Keychain
	AuthConfig(ref reference.Named) registry.AuthConfig
	AuthConfigs() map[string]registry.AuthConfig
}

type AuthProviderConfig struct {
	Logger *zap.SugaredLogger
}

type AuthOpt func(config *AuthProviderConfig)

func WithLogger(logger *zap.Logger) AuthOpt {
	return func(config *AuthProviderConfig) {
		config.Logger = logger.Sugar()
	}
}

func WithSugaredLogger(logger *zap.SugaredLogger) AuthOpt {
	return func(config *AuthProviderConfig) {
		config.Logger = logger
	}
}

func renderConfig(opts []AuthOpt) *AuthProviderConfig {
	config := &AuthProviderConfig{}
	for _, opt := range opts {
		opt(config)
	}
	if config.Logger == nil {
		config.Logger = zap.Must(zap.NewDevelopment()).Sugar()
	}
	return config
}

type SimpleAuthenticator struct {
	AuthConfig registry.AuthConfig
}

func (s *SimpleAuthenticator) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username:      s.AuthConfig.Username,
		Password:      s.AuthConfig.Password,
		Auth:          s.AuthConfig.Auth,
		IdentityToken: s.AuthConfig.IdentityToken,
		RegistryToken: s.AuthConfig.RegistryToken,
	}, nil
}
