package codeRunner

import (
	"code-runner/config"
	errorutil "code-runner/error_util"
	"code-runner/network/wswriter"
	"code-runner/services/container"
	"code-runner/session"
	"context"
)

func (s *Service) Execute(ctx context.Context, id string, params ExecuteParams) error {
	containerConf, containerID, err := s.getContainer(ctx, id, params.SessionKey)
	if err != nil {
		return err
	}
	err = s.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if err != nil {
		return errorutil.ErrorWrap(err, "could not create writer")
	}
	err = s.compile(ctx, containerID, containerConf, params.Writer)
	if err != nil {
		return err
	}
	con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.ExecutionCmd})
	defer con.Close()
	sess, _ := session.GetSession(params.SessionKey)
	sess.Con = con
	err = s.copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		return errorutil.ErrorWrap(err, "execution failed")
	}
	return nil
}

func (s *Service) compile(ctx context.Context, containerID string, containerConf *config.ContainerConfig, writer wswriter.Writer) error {
	if len(containerConf.CompilationCmd) > 0 {
		con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.CompilationCmd})
		if err != nil {
			return errorutil.ErrorWrap(err, "compilation failed")
		}
		defer con.Close()
		s.copy(writer.WithType(wswriter.WriteOutput), con)
	}
	return nil
}
