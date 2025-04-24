package check

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/services/codeRunner"
	"code-runner/session"
	"context"
	"fmt"
	"log"
	"strings"
)

func Check(ctx context.Context, cmdID string, params CheckParams) ([]*model.TestResponseData, error) {
	testResults := make([]*model.TestResponseData, 0)
	containerConf, containerID, err := params.CodeRunner.GetContainer(ctx, cmdID, params.SessionKey)
	if err != nil {
		message := "could not create sandbox environment"
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return nil, errorutil.ErrorWrap(errorSlug, message)
	}
	_, err = params.CodeRunner.GetContainerConnection(ctx, params.SessionKey, containerID, params.Writer, params.RequestId)
	if err != nil {
		message := fmt.Sprintf("could not create connection to container %q", containerID)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return nil, errorutil.ErrorWrap(errorSlug, message)
	}
	err = params.CodeRunner.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if err != nil {
		message := "could not add files to sandbox environment"
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return nil, errorutil.ErrorWrap(errorSlug, message)
	}
	err = params.CodeRunner.Compile(ctx, containerID, containerConf.CompilationCmd, params.Writer)
	if err != nil {
		message := fmt.Sprintf("could not compile program with command %q", containerConf.CompilationCmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return nil, errorutil.ErrorWrap(errorSlug, message)
	}
	sess, err := session.GetSession(params.SessionKey)
	if err != nil {
		message := fmt.Sprintf("could not retreive user session with key %q", params.SessionKey)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return nil, errorutil.ErrorWrap(errorSlug, message)
	}
	var errors = make([]string, 0)
	for _, test := range params.Tests {
		switch test.Type {
		case "output":
			cmd, err := params.CodeRunner.TransformCommand(containerConf.ExecutionCmd, codeRunner.TransformParams{FileName: params.MainFile})
			if err != nil {
				message := fmt.Sprintf("could not execute test of type %s", test.Type)
				errorSlug := errorutil.ErrorSlug()
				log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
				errors = append(errors, errorutil.ErrorWrap(errorSlug, message).Error())
				continue
			}
			testResult, err := outputTest(ctx, sess, cmd, test, params)
			if err != nil {
				message := fmt.Sprintf("could not execute test with command %q", cmd)
				errorSlug := errorutil.ErrorSlug()
				log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
				errors = append(errors, errorutil.ErrorWrap(errorSlug, message).Error())
				continue
			}
			testResults = append(
				testResults,
				testResult,
			)
		case "file":
			var fileCheckParams FileCheckParams
			fileCheckParams.CodeRunner = params.CodeRunner
			fileCheckParams.Writer = params.Writer
			fileCheckParams.ReportPath = containerConf.ReportPath
			fileCheckParams.ReportExtractor = containerConf.ReportExtractor
			cmd, err := params.CodeRunner.TransformCommand(containerConf.ExecutionCmd, codeRunner.TransformParams{FileName: test.Param["filename"]})
			if err != nil {
				message := fmt.Sprintf("could not execute test of type %s", test.Type)
				errorSlug := errorutil.ErrorSlug()
				log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
				errors = append(errors, errorutil.ErrorWrap(errorSlug, message).Error())
				continue
			}
			testResult, err := fileTest(ctx, sess, cmd, containerID, test, fileCheckParams)
			if err != nil {
				message := fmt.Sprintf("could not execute test with command %q", cmd)
				errorSlug := errorutil.ErrorSlug()
				log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
				errors = append(errors, errorutil.ErrorWrap(errorSlug, message).Error())
				continue
			}
			testResults = append(
				testResults,
				testResult,
			)

		}
	}
	if len(errors) > 0 {
		return nil, fmt.Errorf(strings.Join(errors, "\n\n"))
	}
	return testResults, nil
}
