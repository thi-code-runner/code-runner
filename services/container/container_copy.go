package container

import (
	"bytes"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"strings"
	"time"
)

func (cs *Service) CopyToContainer(ctx context.Context, id string, files []*model.SourceFile) error {
	if len(id) <= 0 {
		return fmt.Errorf("could not copy files into docker container because of empty id argument")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if len(files) > 0 {
		var builder strings.Builder
		for i, f := range files {
			buf := make([]byte, base64.StdEncoding.EncodedLen(len(f.Content)))
			base64.StdEncoding.Encode(buf, []byte(f.Content))
			builder.WriteString("echo -n ")
			builder.Write(buf)
			builder.WriteString("| base64 -d >")
			builder.WriteString(f.Filename)
			if i != len(files)-1 {
				builder.WriteString(" && ")
			}
		}
		_, _, err := cs.RunCommand(ctx, id, RunCommandParams{Cmd: builder.String()})
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
		err = cs.cli.CopyToContainer(ctx, id, "/code-runner", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
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
