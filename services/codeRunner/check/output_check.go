package check

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"errors"
	"fmt"
)

func outputTest(ctx context.Context, sess *session.Session, executionCmd string, test *model.TestConfiguration, params CheckParams) (*model.TestResponseData, error) {
	con, _, err := params.CodeRunner.ContainerService.RunCommand(context.Background(), sess.ContainerID, container.RunCommandParams{Cmd: executionCmd, User: "nobody"})
	defer con.Close()
	sess.Con = con
	err = params.CodeRunner.CopyWithTimeout(ctx)(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		message := fmt.Sprintf("could not perform output test with command %q", executionCmd)
		if errors.Is(err, errorutil.TimeoutErr) {
			message = fmt.Sprintf("could not perform output test with command %q, because it timed out", executionCmd)
		}
		return nil, errorutil.ErrorWrap(err, message)
	}
	if string(params.Writer.GetOutput()) == test.Param["expected"] {
		return &model.TestResponseData{Test: test, Passed: true}, nil
	}
	return &model.TestResponseData{Test: test, Message: fmt.Sprintf("output test failed: expected: %q, actual: %q\n", test.Param["expected"], params.Writer.GetOutput()), Passed: false}, nil
}
