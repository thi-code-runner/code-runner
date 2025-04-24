package wswriter

import (
	"bytes"
	"code-runner/model"
	"context"
	"encoding/json"
	"io"

	"nhooyr.io/websocket"
)

const (
	WriteOutput = iota
	WriteError
	WriteTest
	WriteEnd
)

type wsWriterWrapper struct {
	Con *websocket.Conn
}

func (w *wsWriterWrapper) Write(buf []byte) (int, error) {
	return len(buf), w.Con.Write(context.Background(), 1, buf)
}

type Writer interface {
	Write([]byte) (int, error)
	WriteWithSuccess([]byte, bool) ([]byte, error)
	GetOutput() []byte
	WithType(int) Writer
}
type WSWriter struct {
	Con    io.Writer
	Output bytes.Buffer
	Type   int
}

func NewWriter(con *websocket.Conn, t int) *WSWriter {
	return &WSWriter{Con: &wsWriterWrapper{con}, Type: t}
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
		respJson, _ = json.Marshal(resp)
		_, err = ws.Con.Write(respJson)
		ws.Output.Write(buf)
	case WriteError:
		var respJson []byte
		resp := model.ErrorResponse{Type: "output/error", Error: string(buf)}
		respJson, _ = json.Marshal(resp)
		_, err = ws.Con.Write(respJson)
	case WriteTest:
		_, err = ws.Con.Write(buf)
	}
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}

func (ws *WSWriter) WriteWithSuccess(id []byte, success bool) ([]byte, error) {
	var respJson []byte
	resp := model.DoneResponse{Type: "output/end", Id: string(id), Success: success}
	respJson, _ = json.Marshal(resp)
	_, err := ws.Con.Write(respJson)

	if err != nil {
		return nil, err
	}
	return id, nil
}

func (ws *WSWriter) GetOutput() []byte {
	return ws.Output.Bytes()
}
