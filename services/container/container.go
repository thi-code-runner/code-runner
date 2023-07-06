package container

import (
	"context"
	"github.com/docker/docker/client"
	"log"
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
