package podman

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/containers/common/pkg/config"
	"github.com/docker/docker/errdefs"
)

func GetDestination() (config.Connection, error) {
	if runtime.GOOS == "windows" {
		if os.Getenv("TMPDIR") == "" {
			_ = os.Setenv("TMPDIR", "C:\\temp")
		}
	}

	tmpFile, err := os.CreateTemp("", "podman-*.conf")
	if err != nil {
		return config.Connection{}, err
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	_, err = tmpFile.WriteString(fmt.Sprintf("[engine]\ntmp_dir = \"%s\"", strings.ReplaceAll(os.TempDir(), "\\", "\\\\")))
	if err != nil {
		return config.Connection{}, err
	}
	_ = tmpFile.Close()

	conf, err := config.New(&config.Options{
		Modules: []string{tmpFile.Name()},
	})
	if err != nil {
		return config.Connection{}, err
	}
	conns, err := conf.GetAllConnections()
	if err != nil {
		return config.Connection{}, err
	}
	for _, conn := range conns {
		if conn.Default {
			return conn, nil
		}
	}
	return config.Connection{}, errdefs.NotFound(fmt.Errorf("no default connection found"))
}
