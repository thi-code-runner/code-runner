package shell

import "code-runner/services/codeRunner"

type ShellExecuteParams struct {
	SessionKey string
	Stdin      string
	CodeRunner *codeRunner.Service
}
