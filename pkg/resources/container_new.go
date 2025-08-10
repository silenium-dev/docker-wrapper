package resources

import (
	"context"
	"fmt"
	"maps"
	"net"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"k8s.io/utils/ptr"
)

type MountPoint struct {
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
)

type PortMapping struct {
	ContainerPort uint16   `json:"ContainerPort"`
	Protocol      Protocol `json:"Protocol"`
	HostBindings  []HostPortBinding
}

type HostPortBinding struct {
	HostIP   string  `json:"HostIp"`
	HostPort *uint16 `json:"HostPort"`
}

type DnsSpec struct {
	Servers []net.IP `json:"Servers"`
	Search  []string `json:"Search"`
	Options []string `json:"Options"`
}

type CapabilitySpec struct {
	Add  []string `json:"add"`
	Drop []string `json:"drop"`
}

type ContainerSpec struct {
	Labels        map[string]string                     `json:"labels,omitempty"`
	Image         string                                `json:"image,omitempty"`
	Command       []string                              `json:"command,omitempty"`
	Entrypoint    []string                              `json:"entrypoint,omitempty"`
	WorkingDir    string                                `json:"workingdir,omitempty"`
	AutoRemove    bool                                  `json:"autoremove,omitempty"`
	Env           map[string]string                     `json:"env,omitempty"`
	MappedPorts   []PortMapping                         `json:"mappedports,omitempty"`
	ExposedPorts  []nat.Port                            `json:"exposedports,omitempty"`
	Networks      map[*Network]network.EndpointSettings `json:"networks,omitempty"`
	Mounts        []mount.Mount                         `json:"mounts,omitempty"`
	Tty           bool                                  `json:"tty,omitempty"`
	Stdin         bool                                  `json:"stdin,omitempty"`
	Dns           DnsSpec                               `json:"dns,omitempty"`
	Capabilities  CapabilitySpec                        `json:"capabilities,omitempty"`
	ExtraHosts    []string                              `json:"extrahosts,omitempty"`
	HealthCheck   *container.HealthConfig               `json:"healthcheck,omitempty"`
	StopSignal    string                                `json:"stopsignal,omitempty"`
	RestartPolicy container.RestartPolicy               `json:"restartpolicy,omitempty"`
	UsernsMode    container.UsernsMode                  `json:"usernsmode,omitempty"`
	Privileged    bool                                  `json:"privileged,omitempty"`
	LogConfig     container.LogConfig                   `json:"logconfig,omitempty"`
	User          string                                `json:"user,omitempty"`
}

func (c *Client) CreateContainer(
	ctx context.Context, name string, labels ResourceLabels, spec ContainerSpec,
) (*Container, error) {
	allLabels := map[string]string{}
	fullName := name
	if labels != nil {
		maps.Copy(allLabels, labels.ToMap())
		fullName = labels.FullName(name)
	}
	maps.Copy(allLabels, spec.Labels)

	networkConfig := network.NetworkingConfig{
		EndpointsConfig: make(map[string]*network.EndpointSettings),
	}

	for n, cfg := range spec.Networks {
		networkConfig.EndpointsConfig[n.id] = ptr.To(cfg)
	}

	envRaw := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		envRaw = append(envRaw, k+"="+v)
	}

	exposed := make(map[nat.Port]struct{})
	for _, p := range spec.ExposedPorts {
		exposed[p] = struct{}{}
	}

	dnsServers := make([]string, 0, len(spec.Dns.Servers))
	for _, ip := range spec.Dns.Servers {
		dnsServers = append(dnsServers, ip.String())
	}

	portBindings := make(map[nat.Port][]nat.PortBinding)
	for _, p := range spec.MappedPorts {
		port := nat.Port(fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
		hostBindings := make([]nat.PortBinding, 0, len(p.HostBindings))
		for _, h := range p.HostBindings {
			binding := nat.PortBinding{
				HostIP: h.HostIP,
			}
			if h.HostPort != nil {
				binding.HostPort = fmt.Sprintf("%d", *h.HostPort)
			}
			hostBindings = append(hostBindings, binding)
		}

		portBindings[port] = hostBindings
	}

	cfg := &container.CreateRequest{
		Config: &container.Config{
			Labels:       allLabels,
			Image:        spec.Image,
			Cmd:          spec.Command,
			Env:          envRaw,
			Tty:          spec.Tty,
			OpenStdin:    spec.Stdin,
			StopSignal:   spec.StopSignal,
			Healthcheck:  spec.HealthCheck,
			ExposedPorts: exposed,
			Entrypoint:   spec.Entrypoint,
			WorkingDir:   spec.WorkingDir,
			User:         spec.User,
		},
		HostConfig: &container.HostConfig{
			AutoRemove:    spec.AutoRemove,
			Mounts:        spec.Mounts,
			DNS:           dnsServers,
			DNSOptions:    spec.Dns.Options,
			DNSSearch:     spec.Dns.Search,
			ExtraHosts:    spec.ExtraHosts,
			CapAdd:        spec.Capabilities.Add,
			CapDrop:       spec.Capabilities.Drop,
			PortBindings:  portBindings,
			RestartPolicy: spec.RestartPolicy,
			UsernsMode:    spec.UsernsMode,
			Privileged:    spec.Privileged,
			LogConfig:     spec.LogConfig,
		},
		NetworkingConfig: &networkConfig,
	}

	isPodman, err := c.wrapper.SystemIsPodman(ctx)
	if err != nil {
		return nil, err
	}
	if isPodman {
		cfg.HostConfig.LogConfig.Type = "k8s-file"
	}

	resp, err := c.wrapper.ContainerCreate(
		ctx,
		cfg.Config, cfg.HostConfig, cfg.NetworkingConfig,
		nil, fullName,
	)
	if err != nil {
		return nil, err
	}
	return &Container{
		client: c,
		id:     resp.ID,
		name:   name,
	}, nil
}
