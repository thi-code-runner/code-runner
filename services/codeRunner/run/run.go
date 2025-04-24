package run

import (
	errorutil "code-runner/error_util"
	"code-runner/services/codeRunner"
	"context"
	"fmt"
	"log"
)

func Run(ctx context.Context, id string, params ExecuteParams) error {
	containerConf, containerID, err := params.CodeRunner.GetContainer(ctx, id, params.SessionKey)
	if err != nil {
		message := "could not create sandbox environment"
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	con, err := params.CodeRunner.GetContainerConnection(ctx, params.SessionKey, containerID, params.Writer, params.RequestId)
	if err != nil {
		message := fmt.Sprintf("could not create connection to container %q", containerID)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	err = params.CodeRunner.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if err != nil {
		message := "could not add files to sandbox environment"
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	err = params.CodeRunner.Compile(ctx, containerID, containerConf.CompilationCmd, params.Writer)
	if err != nil {
		message := fmt.Sprintf("could not compile program with command %q", containerConf.CompilationCmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	cmd, err := params.CodeRunner.TransformCommand(containerConf.ExecutionCmd, codeRunner.TransformParams{FileName: params.MainFile})
	if err != nil {
		message := fmt.Sprintf("could not execute program %q", params.MainFile)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	err = params.CodeRunner.ContainerService.ExecuteCommand(ctx, con, cmd, true)
	if err != nil {
		message := fmt.Sprintf("could not execute program with command %q", cmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	return nil
}
