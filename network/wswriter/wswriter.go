package wswriter

import (
	"bytes"
	"code-runner/model"
	"context"
	"encoding/json"
	"nhooyr.io/websocket"
)

const (
	WriteOutput = iota
	WriteError
	WriteTest
)

type Writer interface {
	Write([]byte) (int, error)
	GetOutput() []byte
	WithType(int) Writer
}
type WSWriter struct {
	Con    *websocket.Conn
	Output bytes.Buffer
	Type   int
}

func NewWriter(con *websocket.Conn, t int) *WSWriter {
	return &WSWriter{Con: con, Type: t}
}

func (ws *WSWriter) WithType(t int) Writer {
	ws.Type = t
	return ws
}
func (ws *WSWriter) Write(buf []byte) (int, error) {
	var err error
	switch ws.Type {
	case WriteOutput:
		var respJson []byte
		resp := model.RunResponse{Type: "output/run", Data: string(buf)}
		respJson, err = json.Marshal(resp)
		err = ws.Con.Write(context.Background(), 1, respJson)
		ws.Output.Write(buf)
	case WriteError:
		var respJson []byte
		resp := model.ErrorResponse{Type: "output/error", Error: string(buf)}
		respJson, err = json.Marshal(resp)
		err = ws.Con.Write(context.Background(), 1, respJson)
	case WriteTest:
		err = ws.Con.Write(context.Background(), 1, buf)
	}
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}
func (ws *WSWriter) GetOutput() []byte {
	return ws.Output.Bytes()
}
