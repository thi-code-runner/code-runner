package codeRunner

import (
	"code-runner/services/container"
	"context"
)

func (s *Service) Shutdown(ctx context.Context) {
	for _, v := range s.reservedContainers {
		for _, id := range v {
			_ = s.ContainerService.ContainerRemove(ctx, id, container.RemoveCommandParams{Force: true})
		}
	}
	for id := range s.containers {
		_ = s.ContainerService.ContainerRemove(ctx, id, container.RemoveCommandParams{Force: true})
	}
}
