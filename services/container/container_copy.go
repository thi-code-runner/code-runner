package container

import (
	"bytes"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"github.com/docker/docker/api/types"
)

func (cs *Service) CopyToContainer(ctx context.Context, id string, files []*model.SourceFile) error {
	if len(files) > 0 {
		tar, err := createSourceTar(files)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not copy files into container")
		}
		err = cs.cli.CopyToContainer(ctx, id, "/src", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		if err != nil {
			return errorutil.ErrorWrap(err, "could not copy files into container")
		}
	}
	return nil
}
func (cs *Service) CopyResourcesToContainer(ctx context.Context, id string, resources []string) error {
	if len(resources) > 0 {
		tar, err := createTar(resources)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not copy resources into container")
		}
		err = cs.cli.CopyToContainer(ctx, id, "/src", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		if err != nil {
			return errorutil.ErrorWrap(err, "could not copy resources into container")
		}
	}
	return nil
}
