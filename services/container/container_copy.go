package container

import (
	"bytes"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"encoding/base64"
	"fmt"
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
		//gTar, err := gzipTar(tar)
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		buf := make([]byte, base64.RawStdEncoding.EncodedLen(len(tar)))
		base64.RawStdEncoding.Encode(buf, tar)
		cmd := fmt.Sprintf("echo -n %s | base64 -d | tar -xf -", buf)

		// Clean up the container's working directory before copying files
		cleanupCmd := "rm -rf ./*"
		_, _, err = cs.RunCommand(ctx, id, RunCommandParams{Cmd: cleanupCmd})
		if err != nil {
			return errorutil.ErrorWrap(err, fmt.Sprintf("failed to clean up working directory in docker container %q", id))
		}

		_, _, err = cs.RunCommand(ctx, id, RunCommandParams{Cmd: cmd})
		if err != nil {
			return errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files into docker container %q", id))
		}
	}

	time.Sleep(time.Millisecond * 100) // Todo: Replace with a better solution to wait for the container to be ready
	return nil
}
func (cs *Service) CopyFromContainer(ctx context.Context, id string, path string) (string, error) {
	if len(id) <= 0 {
		return "", fmt.Errorf("could not copy files from docker container because of empty id argument")
	}
	cmd := fmt.Sprintf("cat %s", path)
	var buf bytes.Buffer
	con, _, err := cs.RunCommand(context.Background(), id, RunCommandParams{Cmd: cmd})
	defer con.Close()
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files from docker container %q", id))
	}
	_, err = io.Copy(&buf, con)
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not copy files from docker container %q", id))
	}
	return buf.String(), nil
}
