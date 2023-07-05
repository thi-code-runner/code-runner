package codeRunner

import (
	"code-runner/model"
	"context"
	"fmt"
)

func (s *Service) ExecuteCheck(ctx context.Context, cmdID string, params CheckParams) []*model.TestResponse {
	testResults := make([]*model.TestResponse, 0)
	for _, test := range params.Tests {
		switch test.Type {
		case "output":
			testResults = append(
				testResults,
				s.outputTest(ctx, cmdID, test, params),
			)
		}
	}
	return testResults
}
func (s *Service) outputTest(ctx context.Context, cmdID string, test *model.TestConfiguration, params CheckParams) *model.TestResponse {
	s.Execute(ctx, cmdID, ExecuteParams{Writer: params.Writer, SessionKey: params.SessionKey, Files: params.Files, MainFile: params.MainFile})
	if string(params.Writer.GetOutput()) == test.Param["expected"] {
		return &model.TestResponse{Test: test, Passed: true}
	}
	return &model.TestResponse{Test: test, Message: fmt.Sprintf("output test failed: expected: %q, actual: %q\n", test.Param["expected"], params.Writer.GetOutput()), Passed: false}
}
