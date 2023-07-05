package codeRunner

import (
	"code-runner/session"
	"context"
	"fmt"
)

func (s *Service) SendStdIn(ctx context.Context, stdin string, sessionKey string) error {
	var err error
	sess, err := session.GetSession(sessionKey)
	if err != nil || sess.Con == nil {
		return fmt.Errorf("no running execution found to pass stdin to")
	}
	_, err = sess.Con.Write(append([]byte(stdin), '\n'))
	return err
}
