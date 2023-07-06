package container

import (
	errorutil "code-runner/error_util"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
)

func (cs *Service) PullImage(ctx context.Context, img string, w io.Writer) error {
	image, err := cs.cli.ImagePull(ctx, img, types.ImagePullOptions{})
	if err != nil {
		return errorutil.ErrorWrap(err, fmt.Sprintf("docker encountered error while pulling image %q", img))
	}
	defer image.Close()
	io.Copy(w, image)
	return nil
}
