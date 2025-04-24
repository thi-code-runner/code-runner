package run

import (
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner"
)

type ExecuteParams struct {
	Writer     wswriter.Writer
	SessionKey string
	Files      []*model.SourceFile
	MainFile   string
	CodeRunner *codeRunner.Service
	RequestId  string
}
