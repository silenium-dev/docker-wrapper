package resources

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
	"github.com/docker/cli/cli/flags"
	composeapi "github.com/docker/compose/v2/pkg/api"
	compose2 "github.com/docker/compose/v2/pkg/compose"
	"github.com/silenium-dev/docker-wrapper/pkg/api"
	"go.uber.org/zap"
)

type Client struct {
	wrapper       api.ClientWrapper
	dockerCli     *command.DockerCli
	compose       composeapi.Service
	imageProvider ImageProvider

	networks        map[string]*Network
	networksMutex   sync.Mutex
	containers      map[string]*Container
	containersMutex sync.Mutex
	volumes         map[string]*Volume
	volumesMutex    sync.Mutex
}

func NewClient(apiClient api.ClientWrapper, opts ...Opt) (*Client, error) {
	logger := apiClient.Logger()
	stdLog := zap.NewStdLog(logger.Desugar())
	transport := apiClient.HTTPClient().Transport
	skipTlsVerify := false
	if httpTransport, ok := transport.(*http.Transport); ok {
		skipTlsVerify = httpTransport.TLSClientConfig != nil && httpTransport.TLSClientConfig.InsecureSkipVerify
	}
	command.RegisterDefaultStoreEndpoints(
		store.EndpointTypeGetter(
			docker.DockerEndpoint, func() any {
				return &docker.EndpointMeta{
					Host:          apiClient.DaemonHost(),
					SkipTLSVerify: skipTlsVerify,
				}
			},
		),
	)
	dockerCli, err := command.NewDockerCli(
		command.WithAPIClient(apiClient),
		command.WithCombinedStreams(stdLog.Writer()),
		command.WithContentTrustFromEnv(),
		command.WithDefaultContextStoreConfig(),
	)
	if err != nil {
		return nil, err
	}
	err = dockerCli.Initialize(
		&flags.ClientOptions{
			Hosts:     []string{apiClient.DaemonHost()},
			TLSVerify: skipTlsVerify,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker CLI: %w", err)
	}

	compose := compose2.NewComposeService(dockerCli)

	c := &Client{
		wrapper:    apiClient,
		dockerCli:  dockerCli,
		compose:    compose,
		networks:   make(map[string]*Network),
		containers: make(map[string]*Container),
		volumes:    make(map[string]*Volume),
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			_ = c.Close()
			return nil, err
		}
	}
	if c.imageProvider == nil {
		c.imageProvider = DefaultImageProvider()
	}
	return c, nil
}

func (c *Client) Close() error {
	return c.wrapper.Close()
}

var ErrResourceRemoved = errors.New("resource has been removed")

type Opt func(*Client) error

func WithImageProvider(imageProvider ImageProvider) Opt {
	return func(c *Client) error {
		c.imageProvider = imageProvider
		return nil
	}
}
