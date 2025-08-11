package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	client2 "github.com/docker/go-sdk/client"
)

// sdkClient is a type that represents a client for interacting with containers.
type sdkClient struct {
	// log is the logger for the client.
	log *slog.Logger

	// mtx is a mutex for synchronizing access to the fields below.
	mtx sync.RWMutex

	// once is used to initialize the client once.
	once sync.Once

	// client is the underlying docker client, embedded to avoid
	// having to re-implement all the methods.
	dockerClient *client.Client

	// cfg is the configuration for the client, obtained from the environment variables.
	cfg *config

	// err is used to store errors that occur during the client's initialization.
	err error

	// dockerOpts are options to be passed to the docker client.
	dockerOpts []client.Opt

	// dockerContext is the current context of the docker daemon.
	dockerContext string

	// dockerHost is the host of the docker daemon.
	dockerHost string

	// extraHeaders are additional headers to be sent to the docker client.
	extraHeaders map[string]string

	// cached docker info
	dockerInfo    system.Info
	dockerInfoSet bool

	// healthCheck is a function that returns the health of the docker daemon.
	// If not set, the default health check will be used.
	healthCheck func(ctx context.Context) func(c *Client) error
}

// config represents the configuration for the Docker client.
// User values are read from the specified environment variables.
type config struct {
	// Host is the address of the Docker daemon.
	// Default: ""
	Host string `env:"DOCKER_HOST"`

	// TLSVerify is a flag to enable or disable TLS verification when connecting to a Docker daemon.
	// Default: 0
	TLSVerify bool `env:"DOCKER_TLS_VERIFY"`

	// CertPath is the path to the directory containing the Docker certificates.
	// This is used when connecting to a Docker daemon over TLS.
	// Default: ""
	CertPath string `env:"DOCKER_CERT_PATH"`
}

// newConfig returns a new configuration loaded from the properties file
// located in the user's home directory and overridden by environment variables.
func newConfig(host string) (*config, error) {
	cfg := &config{
		Host: host,
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return cfg, nil
}

// validate verifies the configuration is valid.
func (c *config) validate() error {
	if c.TLSVerify && c.CertPath == "" {
		return errors.New("cert path required when TLS is enabled")
	}

	if c.TLSVerify {
		if _, err := os.Stat(c.CertPath); os.IsNotExist(err) {
			return fmt.Errorf("cert path does not exist: %s", c.CertPath)
		}
	}

	if c.Host == "" {
		return errors.New("host is required")
	}

	return nil
}

func (c *Client) SdkClient() *client2.Client {
	return c.sdkClient
}
