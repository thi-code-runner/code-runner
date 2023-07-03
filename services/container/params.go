package container

import (
	"code-runner/model"
)

type RunCommandParams struct {
	Cmd   string
	Files []*model.SourceFile
}
type RemoveCommandParams struct {
	Force bool
}
