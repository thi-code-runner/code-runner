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

func (cs *Service) CreateAndStartContainer(ctx context.Context, image string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
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
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not create container with image %s", image))
	}
	err = cs.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", errorutil.ErrorWrap(err, fmt.Sprintf("could not start container with image %s", image))
	}
	return resp.ID, nil
}
