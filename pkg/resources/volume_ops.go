package resources

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"testing/fstest"

	"github.com/docker/docker/api/types/container"
)

func (v *Volume) AddFS(ctx context.Context, fs fs.FS, options container.CopyToContainerOptions) error {
	err := v.ensureAccessContainer(ctx, v.labels)
	if err != nil {
		return err
	}

	v.mutex.RLock()
	defer v.mutex.RUnlock()

	tarBytes := &bytes.Buffer{}
	tarWriter := tar.NewWriter(tarBytes)
	defer tarWriter.Close()
	err = tarWriter.AddFS(fs)
	if err != nil {
		return err
	}

	return v.client.wrapper.CopyToContainer(ctx, v.accessContainer.id, "/mnt", tarBytes, options)
}

func (v *Volume) AddFiles(
	ctx context.Context, files map[string]*fstest.MapFile, options container.CopyToContainerOptions,
) error {
	contentFs := fstest.MapFS{}
	for path, content := range files {
		contentFs[path] = content
	}
	return v.AddFS(ctx, contentFs, options)
}

func (v *Volume) ReadFile(ctx context.Context, path string) (io.ReadCloser, container.PathStat, error) {
	err := v.ensureAccessContainer(ctx, v.labels)
	if err != nil {
		return nil, container.PathStat{}, err
	}

	v.mutex.RLock()
	defer v.mutex.RUnlock()

	reader, stat, err := v.client.wrapper.CopyFromContainer(ctx, v.accessContainer.id, "/mnt/"+path)
	if err != nil {
		return nil, container.PathStat{}, err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	hdr, err := tarReader.Next()
	if err != nil {
		return nil, container.PathStat{}, err
	}
	if hdr.Size != stat.Size {
		return nil, container.PathStat{}, fmt.Errorf("unexpected size: %d != %d", hdr.Size, stat.Size)
	}
	return &tarReaderWrapper{tarReader, reader}, stat, nil
}

type tarReaderWrapper struct {
	*tar.Reader
	base io.ReadCloser
}

func (t *tarReaderWrapper) Close() error {
	return t.base.Close()
}
