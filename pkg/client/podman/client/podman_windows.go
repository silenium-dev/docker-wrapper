//go:build windows

package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Microsoft/go-winio"
	"github.com/blang/semver/v4"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/version"
	client2 "github.com/silenium-dev/docker-wrapper/pkg/client"
	config2 "github.com/silenium-dev/docker-wrapper/pkg/client/podman/config"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func getConnection(ctx context.Context, cli *client2.Client, logger *zap.SugaredLogger) (config2.Connection, error) {
	// Workaround issue with podman using the temp directory from the machine if no override is provided.
	// /var/tmp is not a valid path for Windows.
	tmpFile, err := os.CreateTemp("", "podman-*.conf")
	if err != nil {
		return config2.Connection{}, err
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	_, err = tmpFile.WriteString(
		fmt.Sprintf(
			"[engine]\ntmp_dir = \"%s\"", strings.ReplaceAll(os.TempDir(), "\\", "\\\\"),
		),
	)
	if err != nil {
		return config2.Connection{}, err
	}
	_ = tmpFile.Close()

	conf, err := config2.New(
		&config2.Options{
			Modules: []string{tmpFile.Name()},
		},
	)
	if err != nil {
		return config2.Connection{}, err
	}

	return deriveConnection(ctx, cli, conf, logger)
}

func directConnection(ctx context.Context, uri *url.URL) (*bindings.Connection, *semver.Version, error) {
	var conn *bindings.Connection
	var ver *semver.Version
	var err error
	if uri.Scheme == "npipe" {
		conn = npipeClient(uri)
		ver, err = pingNewConnection(ctx, conn)
		if err != nil {
			return nil, nil, err
		}
	} else {
		connCtx, err := bindings.NewConnection(ctx, uri.String())
		if err != nil {
			return nil, nil, err
		}
		conn, _ = bindings.GetClient(connCtx)
		ver = bindings.ServiceVersion(connCtx)
	}

	return conn, ver, nil
}

func npipeClient(_url *url.URL) *bindings.Connection {
	connection := &bindings.Connection{URI: _url}
	connection.Client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, _url.Path)
			},
			DisableCompression: true,
		},
	}
	return connection
}

// pingNewConnection pings to make sure the RESTFUL service is up
// and running. it should only be used when initializing a connection
func pingNewConnection(ctx context.Context, conn *bindings.Connection) (*semver.Version, error) {
	// the ping endpoint sits at / in this case
	response, err := conn.DoRequest(ctx, nil, http.MethodGet, "/_ping", nil, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		versionHdr := response.Header.Get("Libpod-API-Version")
		if versionHdr == "" {
			logrus.Warn("Service did not provide Libpod-API-Version Header")
			return new(semver.Version), nil
		}
		versionSrv, err := semver.ParseTolerant(versionHdr)
		if err != nil {
			return nil, err
		}

		switch version.APIVersion[version.Libpod][version.MinimalAPI].Compare(versionSrv) {
		case -1, 0:
			// Server's job when Client version is equal or older
			return &versionSrv, nil
		case 1:
			return nil, fmt.Errorf("server API version is too old. Client %q server %q",
				version.APIVersion[version.Libpod][version.MinimalAPI].String(), versionSrv.String())
		}
	}
	return nil, fmt.Errorf("ping response was %d", response.StatusCode)
}
