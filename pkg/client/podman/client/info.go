package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containers/podman/v5/libpod/define"
)

var ErrSocketNotEnabled = fmt.Errorf("podman socket is not enabled")

// RemoteSocket returns the path to the podman socket on the podman host.
// When talking with a podman machine, this might not be reachable for the caller.
func (p *Podman) RemoteSocket(ctx context.Context) (string, error) {
	info, err := p.SystemInfo(ctx)
	if err != nil {
		return "", err
	}
	if info.Host == nil || info.Host.RemoteSocket == nil || !info.Host.RemoteSocket.Exists {
		return "", ErrSocketNotEnabled
	}
	return info.Host.RemoteSocket.Path, nil
}

func (p *Podman) SystemInfo(ctx context.Context) (define.Info, error) {
	resp, err := p.conn.DoRequest(ctx, nil, "GET", "/info", nil, nil)
	if err != nil {
		return define.Info{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	var info define.Info
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return define.Info{}, err
	}
	return info, nil
}
