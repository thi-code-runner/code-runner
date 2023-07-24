package run

import (
	errorutil "code-runner/error_util"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner"
	"code-runner/services/container"
	"code-runner/session"
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
	con, _, err := params.CodeRunner.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: cmd, User: "nobody"})
	if err != nil {
		message := fmt.Sprintf("could not execute program with command %q", containerConf.ExecutionCmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	defer con.Close()
	sess, err := session.GetSession(params.SessionKey)
	if err != nil {
		message := fmt.Sprintf("could not retreive user session with key %q", params.SessionKey)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	sess.Con = con
	err = params.CodeRunner.Copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		message := fmt.Sprintf("could not send result of compilation with command %q", containerConf.ExecutionCmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	return nil
}
