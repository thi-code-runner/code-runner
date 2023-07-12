package check

import (
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner"
)

type CheckParams struct {
	Writer     wswriter.Writer
	SessionKey string
	Files      []*model.SourceFile
	CodeRunner *codeRunner.Service
	Tests      []*model.TestConfiguration
}
