package container

import (
	"code-runner/model"
)

type RunCommandParams struct {
	Cmd       string
	Files     []*model.SourceFile
	Resources []string
}
type RemoveCommandParams struct {
	Force bool
}
