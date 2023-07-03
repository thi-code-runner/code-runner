package codeRunner

import (
	"code-runner/model"
	"nhooyr.io/websocket"
)

type ExecuteParams struct {
	Con        *websocket.Conn
	SessionKey string
	Files      []*model.SourceFile
	MainFile   string
}
