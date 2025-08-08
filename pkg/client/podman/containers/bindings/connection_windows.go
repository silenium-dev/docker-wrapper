//go:build windows

package bindings

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"github.com/Microsoft/go-winio"
)

func npipeClient(_url *url.URL, _ string, _ string, _ bool) (Connection, error) {
	connection := Connection{URI: _url}
	connection.Client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, _url.Path)
			},
			DisableCompression: true,
		},
	}
	return connection, nil
}

func init() {
	println("registering npipe connection factory")
	connectionFactories["npipe"] = npipeClient
}
