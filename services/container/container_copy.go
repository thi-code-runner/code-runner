package container

import (
	"bytes"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"time"
)

func (cs *Service) CopyToContainer(ctx context.Context, id string, files []*model.SourceFile) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if len(files) > 0 {
		tar, err := createSourceTar(files)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		err = cs.cli.CopyToContainer(ctx, id, "/src", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		if err != nil {
			return errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files into docker container %s", id))
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
			return errorutil.ErrorWrap(err, fmt.Sprintf("could not copy resources into docker container %s", id))
		}
	}
	return nil
}
