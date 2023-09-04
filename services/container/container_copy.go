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
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create tar archive")
		}
		//err = cs.cli.CopyToContainer(ctx, id, "/code-runner", bytes.NewReader(tar), types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(tar)))
		base64.StdEncoding.Encode(buf, tar)

		exec, err := cs.cli.ContainerExecCreate(ctx, id, types.ExecConfig{User: "nobody", AttachStdin: true, AttachStderr: true, AttachStdout: true, Tty: true, WorkingDir: "/code-runner", Cmd: []string{"sh", "-c", "sh"}})
		hijackedResponse, err := cs.cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{Tty: true, Detach: false})
		defer hijackedResponse.Close()
		hijackedResponse.Conn.Write([]byte("echo -n "))
		hijackedResponse.Conn.Write(buf)
		hijackedResponse.Conn.Write(append([]byte(" | base64 -d | tar -xf -"), '\n'))
		io.Copy(os.Stdout, hijackedResponse.Conn)
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
