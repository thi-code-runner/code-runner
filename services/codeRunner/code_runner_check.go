package codeRunner

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"fmt"
)

func (s *Service) ExecuteCheck(ctx context.Context, cmdID string, params CheckParams) ([]*model.TestResponseData, error) {
	testResults := make([]*model.TestResponseData, 0)
	containerConf, containerID, err := s.getContainer(ctx, cmdID, params.SessionKey)
	if err != nil {
		return nil, err
	}
	err = s.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if err != nil {
		return nil, errorutil.ErrorWrap(err, "could not create files")
	}
	err = s.compile(ctx, containerID, containerConf, params.Writer)
	if err != nil {
		return nil, err
	}
	sess, _ := session.GetSession(params.SessionKey)
	for _, test := range params.Tests {
		switch test.Type {
		case "output":
			testResult, _ := s.outputTest(ctx, sess, containerConf.ExecutionCmd, test, params)
			testResults = append(
				testResults,
				testResult,
			)
		case "file":
			testResult, _ := s.fileTest(ctx, sess, containerConf.ExecutionCmd, test, params)
			testResults = append(
				testResults,
				testResult,
			)

		}
	}
	return testResults, nil
}
func (s *Service) outputTest(ctx context.Context, sess *session.Session, executionCmd string, test *model.TestConfiguration, params CheckParams) (*model.TestResponseData, error) {
	con, _, err := s.ContainerService.RunCommand(ctx, sess.ContainerID, container.RunCommandParams{Cmd: executionCmd})
	defer con.Close()
	sess.Con = con
	err = s.copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		return nil, errorutil.ErrorWrap(err, "execution failed")
	}
	if string(params.Writer.GetOutput()) == test.Param["expected"] {
		return &model.TestResponseData{Test: test, Passed: true}, nil
	}
	return &model.TestResponseData{Test: test, Message: fmt.Sprintf("output test failed: expected: %q, actual: %q\n", test.Param["expected"], params.Writer.GetOutput()), Passed: false}, nil
}
func (s *Service) fileTest(ctx context.Context, sess *session.Session, executionCmd string, test *model.TestConfiguration, params CheckParams) (*model.TestResponseData, error) {
	con, executionID, err := s.ContainerService.RunCommand(ctx, sess.ContainerID, container.RunCommandParams{Cmd: executionCmd})
	defer con.Close()
	sess.Con = con
	err = s.copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		return nil, errorutil.ErrorWrap(err, "execution failed")
	}
	code, err := s.ContainerService.GetReturnCode(ctx, executionID)
	if err != nil {
		return nil, err
	}
	if code != 0 {
		return &model.TestResponseData{Test: test, Message: fmt.Sprintf("file test failed with error code %d", code), Passed: false}, nil
	}
	return &model.TestResponseData{Test: test, Message: "", Passed: true}, nil
}
