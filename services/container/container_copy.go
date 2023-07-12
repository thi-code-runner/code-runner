package container

import (
	"bytes"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"time"
)

func (cs *Service) CopyToContainer(ctx context.Context, id string, files []*model.SourceFile) error {
	if len(id) <= 0 {
		return fmt.Errorf("could not copy files into docker container because of empty id argument")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if len(files) > 0 {
		tar, err := createSourceTar(files)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		err = cs.cli.CopyToContainer(ctx, id, "/src", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		if err != nil {
			return errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files into docker container %q", id))
		}
	}
	return nil
}
func (cs *Service) CopyResourcesToContainer(ctx context.Context, id string, resources []string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if len(resources) > 0 {
		tar, err := createTar(resources)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		err = cs.cli.CopyToContainer(ctx, id, "/src", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		if err != nil {
			return errorutil.ErrorWrap(err, fmt.Sprintf("could not copy resources into docker container %q", id))
		}
	}
	return nil
}
func (cs *Service) CopyFromContainer(ctx context.Context, id string, path string) (io.Reader, error) {
	if len(id) <= 0 {
		return nil, fmt.Errorf("could not copy files from docker container because of empty id argument")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	r, _, err := cs.cli.CopyFromContainer(ctx, id, path)
	if err != nil {
		return nil, errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files from docker container %q", id))
	}
	return r, nil
}
