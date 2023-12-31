package server

import (
	"code-runner/config"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner"
	"code-runner/services/codeRunner/check"
	"code-runner/services/codeRunner/input"
	"code-runner/services/codeRunner/run"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

type Server struct {
	mux  *http.ServeMux
	port int
	addr string

	CodeRunner *codeRunner.Service
}

func NewServer(port int, addr string) (*Server, error) {
	if port < 1024 || port > 49151 {
		return nil, errorutil.ErrorWrap(fmt.Errorf("port must be between [1024;49151] but was %d", port), "validation failed")
	}
	if len(addr) <= 0 {
		addr = "localhost"
	}
	return &Server{mux: &http.ServeMux{}, port: port, addr: addr}, nil
}

func (s *Server) Run() {
	fmt.Printf("\n\u2584\u2584\u2584\u2584\u2584\u2584\u2584 \u2584\u2584\u2584\u2584\u2584\u2584\u2584 \u2584\u2584\u2584\u2584\u2584\u2584  \u2584\u2584\u2584\u2584\u2584\u2584\u2584    \u2584\u2584\u2584\u2584\u2584\u2584   \u2584\u2584   \u2584\u2584 \u2584\u2584    \u2584 \u2584\u2584    \u2584 \u2584\u2584\u2584\u2584\u2584\u2584\u2584 \u2584\u2584\u2584\u2584\u2584\u2584   \n\u2588       \u2588       \u2588      \u2588\u2588       \u2588  \u2588   \u2584  \u2588 \u2588  \u2588 \u2588  \u2588  \u2588  \u2588 \u2588  \u2588  \u2588 \u2588       \u2588   \u2584  \u2588  \n\u2588       \u2588   \u2584   \u2588  \u2584    \u2588    \u2584\u2584\u2584\u2588  \u2588  \u2588 \u2588 \u2588 \u2588  \u2588 \u2588  \u2588   \u2588\u2584\u2588 \u2588   \u2588\u2584\u2588 \u2588    \u2584\u2584\u2584\u2588  \u2588 \u2588 \u2588  \n\u2588     \u2584\u2584\u2588  \u2588 \u2588  \u2588 \u2588 \u2588   \u2588   \u2588\u2584\u2584\u2584   \u2588   \u2588\u2584\u2584\u2588\u2584\u2588  \u2588\u2584\u2588  \u2588       \u2588       \u2588   \u2588\u2584\u2584\u2584\u2588   \u2588\u2584\u2584\u2588\u2584 \n\u2588    \u2588  \u2588  \u2588\u2584\u2588  \u2588 \u2588\u2584\u2588   \u2588    \u2584\u2584\u2584\u2588  \u2588    \u2584\u2584  \u2588       \u2588  \u2584    \u2588  \u2584    \u2588    \u2584\u2584\u2584\u2588    \u2584\u2584  \u2588\n\u2588    \u2588\u2584\u2584\u2588       \u2588       \u2588   \u2588\u2584\u2584\u2584   \u2588   \u2588  \u2588 \u2588       \u2588 \u2588 \u2588   \u2588 \u2588 \u2588   \u2588   \u2588\u2584\u2584\u2584\u2588   \u2588  \u2588 \u2588\n\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588  \u2588\u2584\u2584\u2584\u2588  \u2588\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2588  \u2588\u2584\u2584\u2588\u2584\u2588  \u2588\u2584\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2584\u2584\u2588  \u2588\u2584\u2588\n")
	log.Printf("starting code-runner on %s\n", fmt.Sprintf("%s:%d", s.addr, s.port))
	s.initRoutes()
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.addr, s.port), s.mux))
}

func (s *Server) initRoutes() {
	s.mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		var sessionKey string
		sessionKey = uuid.New().String()
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			w.WriteHeader(426)
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")
		for {
			_, buf, err := c.Read(r.Context())
			if err != nil {
				if c.Ping(r.Context()) != nil {
					break
				}
			}
			var v model.Request
			err = json.Unmarshal(buf, &v)
			if err != nil {
				err = errorutil.ErrorWrap(err, "code-runner failed\n\trequest encountered json parse error")
				wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
				log.Println(err)
				continue
			}
			if err = v.Validate(); err != nil {
				err = errorutil.ErrorWrap(err, "code-runner failed\n\trequest validation error")
				wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
				log.Println(err)
				continue
			}
			switch v.Type {
			case "execute/run":
				go func() {
					var runRequest model.RunRequest
					err = json.Unmarshal(buf, &runRequest)
					if err != nil {
						err = errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/run failed\n\trequest encountered json parse error")
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
						log.Println(err)
						return
					}
					if err = runRequest.Validate(); err != nil {
						err = errorutil.ErrorWrap(err, "code-runner failed\n\trequest validation error")
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
						log.Println(err)
						return
					}
					data := runRequest.Data
					if data.Timeout == 0 {
						cconfig := config.GetContainerConfig(data.Cmd)
						if cconfig != nil && cconfig.Timeout != 0 {
							data.Timeout = cconfig.Timeout
						} else {
							data.Timeout = 10
						}
					}
					wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
					ctx, cancel := context.WithTimeout(r.Context(), time.Duration(data.Timeout)*time.Second)
					defer cancel()
					err := run.Run(
						ctx,
						data.Cmd,
						run.ExecuteParams{SessionKey: sessionKey, Writer: wsWriter, Files: data.Sourcefiles, MainFile: data.Mainfilename, CodeRunner: s.CodeRunner},
					)
					if err != nil {
						wsWriter.WithType(wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/run failed").Error()))
						return
					}
				}()
				break
			case "execute/input":
				go func() {
					var stdinRequest model.StdinRequest
					err = json.Unmarshal(buf, &stdinRequest)
					if err != nil {
						err = errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/input failed\n\trequest encountered json parse error")
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
						log.Println(err)
						return
					}
					if err = stdinRequest.Validate(); err != nil {
						err = errorutil.ErrorWrap(err, "code-runner failed\n\trequest validation error")
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
						log.Println(err)
						return
					}
					err := input.Input(r.Context(), stdinRequest.Stdin, sessionKey)
					if err != nil {
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/input failed").Error()))
						return
					}
				}()
			case "execute/test":
				go func() {
					var testRequest model.TestRequest
					err = json.Unmarshal(buf, &testRequest)
					if err != nil {
						err = errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/test failed\n\trequest encountered json parse error")
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
						log.Println(err)
						return
					}
					if err = testRequest.Validate(); err != nil {
						err = errorutil.ErrorWrap(err, "code-runner failed\n\trequest validation error")
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(err.Error()))
						log.Println(err)
						return
					}
					if testRequest.Data.Timeout == 0 {
						cconfig := config.GetContainerConfig(testRequest.Data.Cmd)
						if cconfig != nil && cconfig.Timeout != 0 {
							testRequest.Data.Timeout = cconfig.Timeout
						} else {
							testRequest.Data.Timeout = 10
						}
					}
					wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
					ctx, cancel := context.WithTimeout(r.Context(), time.Duration(testRequest.Data.Timeout)*time.Second)
					defer cancel()
					testResults, err := check.Check(
						ctx,
						testRequest.Data.Cmd,
						check.CheckParams{Writer: wsWriter, SessionKey: sessionKey, MainFile: testRequest.Data.Mainfilename, Files: testRequest.Data.Sourcefiles,
							Tests: testRequest.Data.Tests, CodeRunner: s.CodeRunner},
					)

					if err != nil {
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/test failed").Error()))
						return
					}
					testResult := model.TestResponse{Type: "output/test", Data: testResults}
					testResultJson, err := json.Marshal(testResult)
					if err != nil {
						wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "code-runner failed\n\trexecute/test failed\n\trequest encountered json parse error").Error()))
						log.Println(err)
						return
					}
					wsWriter.WithType(wswriter.WriteTest).Write(testResultJson)
				}()
			default:
				wswriter.NewWriter(c, wswriter.WriteError).Write([]byte("code-runner failed\n\tunrecognized websocket message type"))
			}
		}
		c.Close(websocket.StatusNormalClosure, "")
	})
}
