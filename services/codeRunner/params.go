package codeRunner

import (
	"code-runner/model"
	"code-runner/network/wswriter"
)

type ExecuteParams struct {
	Writer     wswriter.Writer
	SessionKey string
	Files      []*model.SourceFile
	MainFile   string
}

type CheckParams struct {
	ExecuteParams
	Tests []*model.TestConfiguration
}
