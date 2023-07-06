package container

import (
	errorutil "code-runner/error_util"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"net"
)

func (cs *Service) RunCommand(ctx context.Context, id string, params RunCommandParams) (net.Conn, string, error) {

	exec, err := cs.cli.ContainerExecCreate(ctx, id, types.ExecConfig{AttachStdin: true, AttachStderr: true, AttachStdout: true, Tty: true, WorkingDir: "/src", Cmd: []string{"sh", "-c", params.Cmd}})
	if err != nil {
		return nil, "", errorutil.ErrorWrap(err, fmt.Sprintf("docker encountered error_util while executing command %q", params.Cmd))
	}
	hijackedResponse, err := cs.cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{Tty: true, Detach: false})
	return hijackedResponse.Conn, exec.ID, nil
}
