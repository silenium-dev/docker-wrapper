//go:build windows

package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/containers/common/pkg/config"
)

func (c *Client) getConnection() (config.Connection, error) {
	// Workaround issue with podman using the temp directory from the machine if no override is provided.
	// /var/tmp is not a valid path for Windows.
	if os.Getenv("TMPDIR") == "" {
		_ = os.Setenv("TMPDIR", "C:\\temp")
		defer func() { _ = os.Unsetenv("TMPDIR") }()
	}

	tmpFile, err := os.CreateTemp("", "podman-*.conf")
	if err != nil {
		return config.Connection{}, err
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
		return config.Connection{}, err
	}
	_ = tmpFile.Close()

	conf, err := config.New(
		&config.Options{
			Modules: []string{tmpFile.Name()},
		},
	)
	if err != nil {
		return config.Connection{}, err
	}

	return c.getDefaultConnection(conf)
}
