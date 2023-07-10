package container

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"strings"
)

type Service struct {
	cli *client.Client
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

func (cs *Service) GetReturnCode(ctx context.Context, s string) (int, error) {
	execInspect, err := cs.cli.ContainerExecInspect(ctx, s)
	if err != nil {
		return 1, err
	}
	return execInspect.ExitCode, nil
}

func (cs *Service) GetContainers(ctx context.Context) ([]string, error) {
	result := make([]string, 0)
	containerList, err := cs.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	for _, container := range containerList {
		if strings.Contains(container.Names[0], "code-runner-container") {
			result = append(result, container.ID)
		}
		fmt.Println(container.Names[0])
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
