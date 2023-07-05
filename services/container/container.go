package container

import (
	"bytes"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
)

type Service struct {
	cli *client.Client
}

func (cs *Service) RunCommand(ctx context.Context, id string, params RunCommandParams) (net.Conn, string, error) {

	exec, err := cs.cli.ContainerExecCreate(ctx, id, types.ExecConfig{AttachStdin: true, AttachStderr: true, AttachStdout: true, Tty: true, WorkingDir: "/src", Cmd: []string{"sh", "-c", params.Cmd}})
	if err != nil {
		return nil, "", errorutil.ErrorWrap(err, fmt.Sprintf("docker encountered error_util while executing command %q", params.Cmd))
	}
	hijackedResponse, err := cs.cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{Tty: true, Detach: false})
	return hijackedResponse.Conn, exec.ID, nil
}
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
func (cs *Service) PullImage(ctx context.Context, img string, w io.Writer) error {
	image, err := cs.cli.ImagePull(context.Background(), img, types.ImagePullOptions{})
	if err != nil {
		return errorutil.ErrorWrap(err, fmt.Sprintf("docker encountered error while pulling image %q", img))
	}
	defer image.Close()
	io.Copy(w, image)
	return nil
}
func NewService() *Service {
	var cs Service
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("could not connect to docker engine because of %q", err)
	}
	cs.cli = cli
	return &cs
}

func (cs *Service) CreateAndStartContainer(ctx context.Context, image string) (string, error) {
	containerName := fmt.Sprintf("code-runner-container-%s", uuid.New().String())
	resp, err := cs.cli.ContainerCreate(ctx, &container.Config{
		Image:        image,
		Cmd:          []string{"/bin/sh"},
		WorkingDir:   "/src",
		Tty:          true,
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		OpenStdin:    true,
	}, &container.HostConfig{NetworkMode: "none", AutoRemove: true}, nil, nil, containerName)
	err = cs.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (cs *Service) ContainerRemove(ctx context.Context, id string, params RemoveCommandParams) error {
	return cs.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: params.Force})
}
