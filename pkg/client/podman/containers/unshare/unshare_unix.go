//go:build linux

package unshare

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/containers/storage/pkg/idtools"
	"github.com/moby/sys/capability"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

var (
	isRootlessOnce sync.Once
	isRootless     bool
)

const (
	// UsernsEnvName is the environment variable, if set indicates in rootless mode
	UsernsEnvName = "_CONTAINERS_USERNS_CONFIGURED"
)

// hasFullUsersMappings checks whether the current user namespace has all the IDs mapped.
func hasFullUsersMappings() (bool, error) {
	content, err := os.ReadFile("/proc/self/uid_map")
	if err != nil {
		return false, err
	}
	// The kernel rejects attempts to create mappings where either starting
	// point is (u32)-1: https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/kernel/user_namespace.c?id=af3e9579ecfb#n1006 .
	// So, if the uid_map contains 4294967295, the entire IDs space is available in the
	// user namespace, so it is likely the initial user namespace.
	return bytes.Contains(content, []byte("4294967295")), nil
}

var (
	hasCapSysAdminOnce sync.Once
	hasCapSysAdminRet  bool
	hasCapSysAdminErr  error
)

// IsRootless tells us if we are running in rootless mode
func IsRootless() bool {
	isRootlessOnce.Do(func() {
		isRootless = GetRootlessUID() != 0 || os.Getenv(UsernsEnvName) != ""
		if !isRootless {
			hasCapSysAdmin, err := HasCapSysAdmin()
			if err != nil {
				logrus.Warnf("Failed to read CAP_SYS_ADMIN presence for the current process")
			}
			if err == nil && !hasCapSysAdmin {
				isRootless = true
			}
		}
		if !isRootless {
			hasMappings, err := hasFullUsersMappings()
			if err != nil {
				logrus.Warnf("Failed to read current user namespace mappings")
			}
			if err == nil && !hasMappings {
				isRootless = true
			}
		}
	})
	return isRootless
}

type Runnable interface {
	Run() error
}

// getHostIDMappings reads mappings from the named node under /proc.
func getHostIDMappings(path string) ([]specs.LinuxIDMapping, error) {
	var mappings []specs.LinuxIDMapping
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading ID mappings from %q: %w", path, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("line %q from %q has %d fields, not 3", line, path, len(fields))
		}
		cid, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parsing container ID value %q from line %q in %q: %w", fields[0], line, path, err)
		}
		hid, err := strconv.ParseUint(fields[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parsing host ID value %q from line %q in %q: %w", fields[1], line, path, err)
		}
		size, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parsing size value %q from line %q in %q: %w", fields[2], line, path, err)
		}
		mappings = append(mappings, specs.LinuxIDMapping{ContainerID: uint32(cid), HostID: uint32(hid), Size: uint32(size)})
	}
	return mappings, nil
}

// GetHostIDMappings reads mappings for the specified process (or the current
// process if pid is "self" or an empty string) from the kernel.
func GetHostIDMappings(pid string) ([]specs.LinuxIDMapping, []specs.LinuxIDMapping, error) {
	if pid == "" {
		pid = "self"
	}
	uidmap, err := getHostIDMappings(fmt.Sprintf("/proc/%s/uid_map", pid))
	if err != nil {
		return nil, nil, err
	}
	gidmap, err := getHostIDMappings(fmt.Sprintf("/proc/%s/gid_map", pid))
	if err != nil {
		return nil, nil, err
	}
	return uidmap, gidmap, nil
}

// GetSubIDMappings reads mappings from /etc/subuid and /etc/subgid.
func GetSubIDMappings(user, group string) ([]specs.LinuxIDMapping, []specs.LinuxIDMapping, error) {
	mappings, err := idtools.NewIDMappings(user, group)
	if err != nil {
		return nil, nil, fmt.Errorf("reading subuid mappings for user %q and subgid mappings for group %q: %w", user, group, err)
	}
	var uidmap, gidmap []specs.LinuxIDMapping
	for _, m := range mappings.UIDs() {
		uidmap = append(uidmap, specs.LinuxIDMapping{
			ContainerID: uint32(m.ContainerID),
			HostID:      uint32(m.HostID),
			Size:        uint32(m.Size),
		})
	}
	for _, m := range mappings.GIDs() {
		gidmap = append(gidmap, specs.LinuxIDMapping{
			ContainerID: uint32(m.ContainerID),
			HostID:      uint32(m.HostID),
			Size:        uint32(m.Size),
		})
	}
	return uidmap, gidmap, nil
}

// ParseIDMappings parses mapping triples.
func ParseIDMappings(uidmap, gidmap []string) ([]idtools.IDMap, []idtools.IDMap, error) {
	uid, err := idtools.ParseIDMap(uidmap, "userns-uid-map")
	if err != nil {
		return nil, nil, err
	}
	gid, err := idtools.ParseIDMap(gidmap, "userns-gid-map")
	if err != nil {
		return nil, nil, err
	}
	return uid, gid, nil
}

// HasCapSysAdmin returns whether the current process has CAP_SYS_ADMIN.
func HasCapSysAdmin() (bool, error) {
	hasCapSysAdminOnce.Do(func() {
		currentCaps, err := capability.NewPid2(0)
		if err != nil {
			hasCapSysAdminErr = err
			return
		}
		if err = currentCaps.Load(); err != nil {
			hasCapSysAdminErr = err
			return
		}
		hasCapSysAdminRet = currentCaps.Get(capability.EFFECTIVE, capability.CAP_SYS_ADMIN)
	})
	return hasCapSysAdminRet, hasCapSysAdminErr
}

// GetRootlessUID returns the UID of the user in the parent userNS
func GetRootlessUID() int {
	uidEnv := os.Getenv("_CONTAINERS_ROOTLESS_UID")
	if uidEnv != "" {
		u, _ := strconv.Atoi(uidEnv)
		return u
	}
	return os.Getuid()
}

// GetRootlessGID returns the GID of the user in the parent userNS
func GetRootlessGID() int {
	gidEnv := os.Getenv("_CONTAINERS_ROOTLESS_GID")
	if gidEnv != "" {
		u, _ := strconv.Atoi(gidEnv)
		return u
	}
	return os.Getgid()
}
