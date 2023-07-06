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
		tar, err := createTar(files)
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
