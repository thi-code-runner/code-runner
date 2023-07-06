package container

import (
	"context"
	"github.com/docker/docker/api/types"
)

func (cs *Service) ContainerRemove(ctx context.Context, id string, params RemoveCommandParams) error {
	return cs.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: params.Force})
}
