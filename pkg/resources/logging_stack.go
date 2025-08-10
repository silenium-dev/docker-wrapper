package resources

import (
	"context"
	"fmt"
	"strings"
	"testing/fstest"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/silenium-dev/docker-wrapper/pkg/client/podman/client"
	"github.com/silenium-dev/docker-wrapper/pkg/errors"
	"github.com/silenium-dev/docker-wrapper/pkg/template"
	"k8s.io/utils/ptr"
)

type LoggingStack struct {
	client        *Client
	project       *types.Project
	lokiVolume    *Volume
	alloyVolume   *Volume
	grafanaVolume *Volume
	labelFilter   map[string]*string
}

const grafanaConfigVolumeName = "logging-stack-grafana-config"
const lokiConfigVolumeName = "logging-stack-loki-config"
const alloyConfigVolumeName = "logging-stack-alloy-config"

type loggingStackLabels struct {
}

func (l loggingStackLabels) ToMap() map[string]string {
	return map[string]string{
		"dev.silenium.docker-wrapper.logging-stack": "true",
	}
}

func (l loggingStackLabels) ToFilter() filters.Args {
	selector := l.ToMap()
	result := strings.Builder{}
	for k, v := range selector {
		if result.Len() > 0 {
			result.WriteString(",")
		}
		result.WriteString(k)
		result.WriteString("=")
		result.WriteString(v)
	}
	return filters.NewArgs(filters.Arg("label", result.String()))
}

func (l loggingStackLabels) FullName(name string) string {
	return fmt.Sprintf("logging-stack-%s", name)
}

func (l loggingStackLabels) TrimName(name string) string {
	return strings.TrimPrefix(name, "logging-stack-")
}

func (c *Client) GetOrCreateLoggingStack(
	ctx context.Context, scrapeLabels map[string]*string, quiet bool,
) (*LoggingStack, error) {
	if scrapeLabels == nil {
		scrapeLabels = map[string]*string{
			"alloy-scrape": ptr.To("true"),
		}
	}

	grafanaConfigVolume, err := c.createOrImportVolumeWithFiles(
		ctx, grafanaConfigVolumeName, map[string]*fstest.MapFile{
			"datasources/loki.yaml": {Data: []byte(grafanaLokiDatasource), Mode: 0o644},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add files to grafana config volume: %w", err)
	}
	lokiConfigVolume, err := c.createOrImportVolumeWithFiles(
		ctx, lokiConfigVolumeName, map[string]*fstest.MapFile{
			"config.yaml": {Data: []byte(lokiConfig), Mode: 0o644},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add files to loki config volume: %w", err)
	}
	isPodman, err := c.wrapper.SystemIsPodman(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if client is podman: %w", err)
	}
	var labelStrings []string
	for k, v := range scrapeLabels {
		suffix := ""
		if v != nil {
			suffix = "=" + *v
		}
		labelStrings = append(labelStrings, fmt.Sprintf("\"%s%s\"", k, suffix))
	}
	renderedAlloyMain, err := template.Render(
		"main.alloy", alloyMainConfig, map[string]any{
			"is_podman": isPodman,
			"labels":    strings.Join(labelStrings, ", "),
		}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to render alloy main config: %w", err)
	}
	alloyConfigVolume, err := c.createOrImportVolumeWithFiles(
		ctx, alloyConfigVolumeName, map[string]*fstest.MapFile{
			"config.alloy": {Data: []byte(alloyConfig), Mode: 0o644},
			"main.alloy":   {Data: []byte(renderedAlloyMain), Mode: 0o644},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add files to alloy config volume: %w", err)
	}

	project, err := loader.LoadWithContext(
		ctx, types.ConfigDetails{
			ConfigFiles: []types.ConfigFile{
				{
					Filename: "docker-compose.yaml",
					Content:  []byte(composeString),
				},
			},
		}, func(options *loader.Options) {
			options.ResolvePaths = true
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load compose project: %w", err)
	}
	for k, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}
		project.Services[k] = s
	}
	alloyService, err := project.GetService("alloy")
	if err != nil {
		return nil, fmt.Errorf("failed to get alloy service from compose file: %w", err)
	}
	if isPodman {
		podmanConn, err := client.FromDocker(ctx, c.wrapper)
		if err != nil {
			return nil, fmt.Errorf("failed to get podman connection: %w", err)
		}
		podmanInfo, err := podmanConn.SystemInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get podman info: %w", err)
		}
		graphRoot := podmanInfo.Store.GraphRoot

		socketPath, err := podmanConn.RemoteSocket(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get podman socket path: %w", err)
		}

		alloyService.Volumes = append(
			alloyService.Volumes,
			types.ServiceVolumeConfig{
				Type:        "bind",
				Source:      graphRoot,
				Target:      "/var/lib/containers/storage",
				ReadOnly:    true,
				Consistency: "consistent",
			},
			types.ServiceVolumeConfig{
				Type:   "bind",
				Source: strings.TrimPrefix(socketPath, "unix://"),
				Target: "/var/run/docker.sock",
			},
		)
	} else {
		alloyService.Volumes = append(
			alloyService.Volumes,
			types.ServiceVolumeConfig{
				Type:   "bind",
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		)
	}
	project.Services[alloyService.Name] = alloyService

	lokiConfigSpec := project.Volumes["loki-config"]
	lokiConfigSpec.Name = lokiConfigVolume.Name()
	project.Volumes["loki-config"] = lokiConfigSpec

	alloyConfigSpec := project.Volumes["alloy-config"]
	alloyConfigSpec.Name = alloyConfigVolume.Name()
	project.Volumes["alloy-config"] = alloyConfigSpec

	grafanaConfigSpec := project.Volumes["grafana-config"]
	grafanaConfigSpec.Name = grafanaConfigVolume.Name()
	project.Volumes["grafana-config"] = grafanaConfigSpec

	err = c.compose.Up(
		ctx, project, api.UpOptions{
			Create: api.CreateOptions{
				QuietPull:     quiet,
				AssumeYes:     true,
				RemoveOrphans: true,
			},
			Start: api.StartOptions{
				Wait: true,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to up logging stack: %w", err)
	}

	return &LoggingStack{
		client:        c,
		alloyVolume:   alloyConfigVolume,
		grafanaVolume: grafanaConfigVolume,
		lokiVolume:    lokiConfigVolume,
		project:       project,
		labelFilter:   scrapeLabels,
	}, nil
}

func (l *LoggingStack) Remove(ctx context.Context, options api.DownOptions) error {
	return l.client.compose.Down(ctx, l.project.Name, options)
}

func (c *Client) createOrImportVolumeWithFiles(ctx context.Context, name string, files map[string]*fstest.MapFile) (
	*Volume, error,
) {
	volume, err := c.ImportVolume(ctx, name)
	if errors.IsNotFound(err, errors.ResourceTypeContainer) {
		volume, err = c.CreateVolume(ctx, loggingStackLabels{}, name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create or import volume %s: %w", name, err)
	}
	err = volume.AddFiles(ctx, files, container.CopyToContainerOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to add files to volume %s: %w", name, err)
	}
	return volume, nil
}
