package request

import (
	errorutil "code-runner/error_util"
	"code-runner/network/wswriter"
	"encoding/json"
	"log"

	"nhooyr.io/websocket"
)

func ParseAndValidateRequest(buf []byte, req interface{}, c *websocket.Conn) error {
	if err := json.Unmarshal(buf, req); err != nil {
		err = errorutil.ErrorWrap(err, "JSON parse error")
		wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
		log.Println(err)
		return err
	}
	if validator, ok := req.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			err = errorutil.ErrorWrap(err, "Request validation error")
			wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
			log.Println(err)
			return err
		}
	}
	return nil
}
