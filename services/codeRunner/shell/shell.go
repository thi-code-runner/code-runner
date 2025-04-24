package shell

import (
	errorutil "code-runner/error_util"
	"code-runner/session"
	"context"
	"fmt"
	"log"
)

func ShellExecute(ctx context.Context, params ShellExecuteParams) error {
	sess, err := session.GetSession(params.SessionKey)
	if err != nil || sess.Con == nil {
		return fmt.Errorf("could not retrieve session with key %q", params.SessionKey)
	}

	err = params.CodeRunner.ContainerService.ExecuteCommand(ctx, sess.Con, params.Stdin, false)
	if err != nil {
		message := fmt.Sprintf("could not execute program with command %q", params.Stdin)
		errorSlug := errorutil.ErrorSlug()
		log.Println(errorutil.ErrorWrap(errorSlug, errorutil.ErrorWrap(err, message).Error()))
		return errorutil.ErrorWrap(errorSlug, message)
	}

	return nil
}
