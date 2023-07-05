package server

import (
	"bytes"
	"code-runner/model"
	"context"
	"encoding/json"
	"io"
	"nhooyr.io/websocket"
)

type WSWriter struct {
	Con    *websocket.Conn
	Data   interface{}
	Output bytes.Buffer
}

type WSOutputWriter struct {
	WSW *WSWriter
}

type WSTestWriter struct {
	WSW *WSWriter
}

func (ws *WSWriter) GetOutputWriter() io.Writer {
	return &WSOutputWriter{WSW: ws}
}
func (ws *WSWriter) GetTestWriter() io.Writer {
	return &WSTestWriter{WSW: ws}
}
func (ws *WSOutputWriter) Write(buf []byte) (int, error) {
	var respJson []byte
	var err error
	resp := model.RunResponse{Output: string(buf)}
	respJson, err = json.Marshal(resp)
	if err != nil {
		return 0, err
	}
	ws.WSW.Output.Write(buf)
	err = ws.WSW.Con.Write(context.Background(), 1, respJson)
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}
func (ws *WSTestWriter) Write(buf []byte) (int, error) {
	var err error
	ws.WSW.Output.Write(buf)
	err = ws.WSW.Con.Write(context.Background(), 1, buf)
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}

func (ws *WSWriter) GetOutput() []byte {
	return ws.Output.Bytes()
}
