package check

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"fmt"
)

func fileTest(ctx context.Context, sess *session.Session, executionCmd string, test *model.TestConfiguration, params CheckParams) (*model.TestResponseData, error) {
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
		return &model.TestResponseData{Test: test, Message: fmt.Sprintf("file test failed with error code %d", code), Passed: false}, nil
	}
	return &model.TestResponseData{Test: test, Message: "", Passed: true}, nil
}
