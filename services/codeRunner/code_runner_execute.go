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
	var containerConf config.ContainerConfig
	for _, c := range config.Conf.ContainerConfig {
		if id == c.ID {
			containerConf = c
			break
		}
	}
	var containerID string
	sess, err := session.GetSession(params.SessionKey)
	if err == nil {
		containerID = sess.ContainerID
	}
	containerID, err = s.getContainerID(ctx, containerID, containerConf)
	if err != nil {
		return err
	}
	sess = session.PutSession(params.SessionKey, &session.Session{ContainerID: containerID})
	err = s.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if len(containerConf.CompilationCmd) > 0 {
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create writer")
		}
		con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.CompilationCmd})
		defer con.Close()
		s.copy(params.Writer.WithType(wswriter.WriteOutput), con)
		if err != nil {
			return errorutil.ErrorWrap(err, "compilation failed")
		}
	}
	if err != nil {
		return errorutil.ErrorWrap(err, "could not create writer")
	}
	con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.ExecutionCmd})
	defer con.Close()
	sess.Con = con
	err = s.copy(params.Writer.WithType(wswriter.WriteOutput), con)
	if err != nil {
		return errorutil.ErrorWrap(err, "execution failed")
	}
	return nil
}
