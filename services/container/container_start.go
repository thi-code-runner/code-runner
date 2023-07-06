package container

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
)

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
