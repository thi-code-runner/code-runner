package container

type RunCommandParams struct {
	CmdID string
}

type RunCommandWithFilesParams struct {
	CmdID string
	Files map[string]string
}
