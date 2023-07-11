package codeRunner

import (
	"code-runner/config"
	errorutil "code-runner/error_util"
	"code-runner/network/wswriter"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"fmt"
	"log"
)

func (s *Service) Execute(ctx context.Context, id string, params ExecuteParams) error {
	containerConf, containerID, err := s.getContainer(ctx, id, params.SessionKey)
	if err != nil {
		message := "could not create sandbox environment"
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	err = s.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if err != nil {
		message := "could not add files to sandbox environment"
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	err = s.compile(ctx, containerID, containerConf, params.Writer)
	if err != nil {
		message := fmt.Sprintf("could not compile program with command %q", containerConf.CompilationCmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.ExecutionCmd})
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
	err = s.copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		message := fmt.Sprintf("could not send result of compilation with command %q", containerConf.ExecutionCmd)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}
	return nil
}

func (s *Service) compile(ctx context.Context, containerID string, containerConf *config.ContainerConfig, writer wswriter.Writer) error {
	if len(containerConf.CompilationCmd) > 0 {
		con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.CompilationCmd})
		if err != nil {
			return err
		}
		defer con.Close()
		err = s.copy(writer.WithType(wswriter.WriteOutput), con)
		if err != nil {
			return err
		}
	}
	return nil
}
