package container

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
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
		//gTar, err := gzipTar(tar)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		buf := make([]byte, base64.RawStdEncoding.EncodedLen(len(tar)))
		base64.RawStdEncoding.Encode(buf, tar)
		cmd := fmt.Sprintf("echo -n %s | base64 -d | tar -xf -", buf)
		_, _, err = cs.RunCommand(ctx, id, RunCommandParams{Cmd: cmd})
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
		gTar, err := gzipTar(tar)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(gTar)))
		base64.StdEncoding.Encode(buf, gTar)
		cmd := fmt.Sprintf("echo -n %s | base64 -d | tar -zxf -", buf)
		c, _, err := cs.RunCommand(ctx, id, RunCommandParams{Cmd: cmd})
		io.Copy(os.Stdout, c)
		if err != nil {
			return errorutil.ErrorWrap(err, fmt.Sprintf("could not copy resources into docker container %q", id))
		}
	}
	return nil
}
func (cs *Service) CopyFromContainer(ctx context.Context, id string, path string) (string, error) {
	if len(id) <= 0 {
		return "", fmt.Errorf("could not copy files from docker container because of empty id argument")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	r, _, err := cs.cli.CopyFromContainer(ctx, id, path)
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files from docker container %q", id))
	}
	defer r.Close()
	result, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
