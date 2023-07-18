package check

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner/check/extractor"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"fmt"
)

func fileTest(ctx context.Context, sess *session.Session, executionCmd string, containerID string, test *model.TestConfiguration, params FileCheckParams) (*model.TestResponseData, error) {
	var resultData model.TestResponseData
	resultData.Test = test
	resultData.Passed = true
	con, executionID, err := params.CodeRunner.ContainerService.RunCommand(ctx, sess.ContainerID, container.RunCommandParams{Cmd: executionCmd})
	defer con.Close()
	sess.Con = con
	err = params.CodeRunner.Copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		return nil, errorutil.ErrorWrap(err, "execution failed")
	}
	code, err := params.CodeRunner.ContainerService.GetReturnCode(ctx, executionID)
	if err != nil {
		return nil, err
	}
	if code != 0 {
		resultData.Message = fmt.Sprintf("file test failed with error code %d", code)
		resultData.Passed = false
	}
	if len(params.ReportPath) > 0 {
		report, err := params.CodeRunner.ContainerService.CopyFromContainer(ctx, containerID, params.ReportPath)
		if err != nil {
			return nil, err
		}
		resultData.Detail = extractor.Extract(params.ReportExtractor, report)
	}

	return &resultData, nil
}
