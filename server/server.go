package server

import (
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"nhooyr.io/websocket"
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
		ses, err := r.Cookie("code-runner-session")
		if err != nil {
			sessionKey = uuid.New().String()
			w.Header().Set("Set-Cookie", fmt.Sprintf("code-runner-session=%s", sessionKey))
		} else {
			sessionKey = ses.Value
		}
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			w.WriteHeader(426)
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")
		for {
			var v model.GenericRequest
			_, buf, err := c.Read(r.Context())
			if err != nil {
				if c.Ping(r.Context()) != nil {
					break
				}
			}
			err = json.Unmarshal(buf, &v)
			if err != nil {
				err = errorutil.ErrorWrap(err, "code-runner failed\n\trequest encountered json parse error")
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
					data := runRequest.Data
					wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
					err := s.CodeRunner.Execute(
						r.Context(),
						data.Cmd,
						codeRunner.ExecuteParams{SessionKey: sessionKey, Writer: wsWriter, Files: data.Sourcefiles, MainFile: data.Mainfilename},
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
					err := s.CodeRunner.SendStdIn(r.Context(), stdinRequest.Stdin, sessionKey)
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
					wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
					testResults, err := s.CodeRunner.ExecuteCheck(
						r.Context(),
						testRequest.Data.Cmd,
						codeRunner.CheckParams{ExecuteParams: codeRunner.ExecuteParams{Writer: wsWriter, SessionKey: sessionKey, Files: testRequest.Data.Sourcefiles},
							Tests: testRequest.Data.Tests},
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
