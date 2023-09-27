package container

import (
	errorutil "code-runner/error_util"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"time"
)

func (cs *Service) CreateAndStartContainer(ctx context.Context, image string, params ContainerCreateParams) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	containerName := fmt.Sprintf("code-runner-container-%s", uuid.New().String())
	var pidsLimit int64 = 100
	var diskSize = "10m"
	if len(params.DiskSize) > 0 {
		diskSize = params.DiskSize
	}
	var resp, err = cs.cli.ContainerCreate(ctx, &container.Config{
		Image:           image,
		Cmd:             []string{"/bin/sh"},
		WorkingDir:      "/code-runner",
		NetworkDisabled: true,
		Tty:             true,
		AttachStderr:    true,
		AttachStdout:    true,
		AttachStdin:     true,
		OpenStdin:       true,
	}, &container.HostConfig{ReadonlyRootfs: true, Tmpfs: map[string]string{"/code-runner": fmt.Sprintf("rw,size=%s,exec", diskSize)}, NetworkMode: "none", AutoRemove: true, Resources: container.Resources{PidsLimit: &pidsLimit, Memory: params.Memory * 1024 * 1024, NanoCPUs: int64(params.CPU * 100000 * 10000)}}, nil, nil, containerName)
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not create container with image %q", image))
	}
	err = cs.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if !params.ReadOnly {
		cs.RunCommand(ctx, resp.ID, RunCommandParams{Cmd: "chmod o+w /code-runner"})
	}
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not start container with image %s", image))
	}
	return resp.ID, nil
}
