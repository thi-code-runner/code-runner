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

func outputTest(ctx context.Context, sess *session.Session, executionCmd string, test *model.TestConfiguration, params CheckParams) (*model.TestResponseData, error) {
	con, _, err := params.CodeRunner.ContainerService.RunCommand(ctx, sess.ContainerID, container.RunCommandParams{Cmd: executionCmd, User: "nobody"})
	defer con.Close()
	sess.Con = con
	err = params.CodeRunner.Copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		return nil, errorutil.ErrorWrap(err, "execution failed")
	}
	if string(params.Writer.GetOutput()) == test.Param["expected"] {
		return &model.TestResponseData{Test: test, Passed: true}, nil
	}
	return &model.TestResponseData{Test: test, Message: fmt.Sprintf("output test failed: expected: %q, actual: %q\n", test.Param["expected"], params.Writer.GetOutput()), Passed: false}, nil
}
