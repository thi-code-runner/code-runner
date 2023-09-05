package container

import (
	"code-runner/model"
)

type RunCommandParams struct {
	Cmd       string
	Memory    int64
	CPU       int64
	Files     []*model.SourceFile
	Resources []string
	User      string
}
type ContainerCreateParams struct {
	Memory   int64
	CPU      float32
	ReadOnly bool
	DiskSize string
}
type RemoveCommandParams struct {
	Force bool
}
