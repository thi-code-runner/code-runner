package codeRunner

import (
	"code-runner/model"
	"io"
)

type Writer interface {
	GetOutputWriter() io.Writer
	GetTestWriter() io.Writer
	GetOutput() []byte
}
type ExecuteParams struct {
	Writer     Writer
	SessionKey string
	Files      []*model.SourceFile
	MainFile   string
}

type CheckParams struct {
	ExecuteParams
	Tests []*model.TestConfiguration
}
