package auth

import (
	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/registry"
	"go.uber.org/zap"
)

type Provider interface {
	AuthConfig(ref reference.Named) registry.AuthConfig
}

type ProviderConfig struct {
	Logger *zap.SugaredLogger
}

type Opt func(config *ProviderConfig)

func WithLogger(logger *zap.Logger) Opt {
	return func(config *ProviderConfig) {
		config.Logger = logger.Sugar()
	}
}

func WithSugaredLogger(logger *zap.SugaredLogger) Opt {
	return func(config *ProviderConfig) {
		config.Logger = logger
	}
}

func renderConfig(opts []Opt) *ProviderConfig {
	config := &ProviderConfig{}
	for _, opt := range opts {
		opt(config)
	}
	if config.Logger == nil {
		config.Logger = zap.Must(zap.NewDevelopment()).Sugar()
	}
	return config
}
