package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/distribution/reference"
	"net/http"
)

func (c *Client) RequestAuthenticate(req *http.Request, ref reference.Named) error {
	authConfig := c.authProvider.AuthConfig(ref)
	confBytes, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}
	encoded := base64.URLEncoding.EncodeToString(confBytes)
	req.Header.Set("X-Registry-Auth", encoded)
	return nil
}
