package codeRunner

import (
	"code-runner/services/container"
	"context"
)

type ContainerService interface {
	RunCommand(context.Context, string, container.RunCommandParams) error
}
type Service struct {
	ContainerService ContainerService
}

func NewService(containerService ContainerService) *Service {
	return &Service{ContainerService: containerService}
}
func (s *Service) Execute(ctx context.Context) {
	s.ContainerService.RunCommand(ctx, "", container.RunCommandParams{CmdID: "java-20"})
}
