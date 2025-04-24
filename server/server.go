package server

import (
	"code-runner/config"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/request"
	"code-runner/network/wswriter"
	"code-runner/services/codeRunner"
	"code-runner/services/codeRunner/check"
	"code-runner/services/codeRunner/run"
	"code-runner/services/codeRunner/shell"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

const (
	defaultTimeout = 10 // Default timeout in seconds
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
	if len(addr) == 0 {
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
	s.mux.HandleFunc("/run", s.handleRun)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	sessionKey := uuid.New().String()
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to establish websocket connection", http.StatusUpgradeRequired)
		return
	}
	defer s.cleanupSession(r.Context(), sessionKey)

	for {
		_, buf, err := c.Read(r.Context())
		if err != nil {
			if c.Ping(r.Context()) != nil {
				break
			}
			continue
		}

		var req model.Request
		if err := request.ParseAndValidateRequest(buf, &req, c); err != nil {
			continue
		}

		switch req.Type {
		case "execute/run":
			go s.handleExecuteRun(r.Context(), buf, sessionKey, c, req.Id)
		case "execute/test":
			go s.handleExecuteTest(r.Context(), buf, sessionKey, c, req.Id)
		case "execute/shell":
			go s.handleExecuteShell(r.Context(), buf, sessionKey, c)
		default:
			wswriter.NewWriter(c, wswriter.WriteError).Write([]byte("Unrecognized websocket message type"))
		}
	}
}

func (s *Server) cleanupSession(ctx context.Context, sessionKey string) {
	sess, err := session.GetSession(sessionKey)
	if err != nil {
		log.Println(errorutil.ErrorWrap(err, "Failed to retrieve session during cleanup"))
	} else {
		session.StopSession(sessionKey)
		session.DeleteSession(sessionKey)
		if err := s.CodeRunner.ContainerService.ContainerRemove(ctx, sess.ContainerID, container.RemoveCommandParams{Force: true}); err != nil {
			log.Println(errorutil.ErrorWrap(err, "Failed to remove container during cleanup"))
		}
	}
}

func (s *Server) handleExecuteRun(ctx context.Context, buf []byte, sessionKey string, c *websocket.Conn, rId string) {
	var runRequest model.RunRequest
	if err := request.ParseAndValidateRequest(buf, &runRequest, c); err != nil {
		wswriter.NewWriter(c, wswriter.WriteEnd).WriteWithSuccess([]byte(rId), false)
		return
	}

	data := runRequest.Data
	if data.Timeout == 0 {
		data.Timeout = s.getDefaultTimeout(data.Cmd)
	}

	wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(data.Timeout)*time.Second)
	defer cancel()

	err := run.Run(ctx, data.Cmd, run.ExecuteParams{
		SessionKey: sessionKey,
		Writer:     wsWriter,
		Files:      data.Sourcefiles,
		MainFile:   data.Mainfilename,
		CodeRunner: s.CodeRunner,
		RequestId:  rId,
	})
	if err != nil {
		wsWriter.WithType(wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "Execute/run failed").Error()))
		wsWriter.WriteWithSuccess([]byte(rId), false)
	}

}

func (s *Server) handleExecuteTest(ctx context.Context, buf []byte, sessionKey string, c *websocket.Conn, rId string) {
	var testRequest model.TestRequest
	if err := request.ParseAndValidateRequest(buf, &testRequest, c); err != nil {
		return
	}

	data := testRequest.Data
	if data.Timeout == 0 {
		data.Timeout = s.getDefaultTimeout(data.Cmd)
	}

	wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(data.Timeout)*time.Second)
	defer cancel()

	testResults, err := check.Check(ctx, data.Cmd, check.CheckParams{
		Writer:     wsWriter,
		SessionKey: sessionKey,
		MainFile:   data.Mainfilename,
		Files:      data.Sourcefiles,
		Tests:      data.Tests,
		CodeRunner: s.CodeRunner,
		RequestId:  rId,
	})
	if err != nil {
		wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "Execute/test failed").Error()))
		wswriter.NewWriter(c, wswriter.WriteEnd).WriteWithSuccess([]byte(rId), false)
		return
	}

	testResult := model.TestResponse{Type: "output/test", Data: testResults}
	testResultJSON, err := json.Marshal(testResult)
	if err != nil {
		wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "JSON marshal error").Error()))
		log.Println(err)
		return
	}

	wsWriter.WithType(wswriter.WriteTest).Write(testResultJSON)

	testsPassed := true
	for _, result := range testResults {
		if !result.Passed {
			testsPassed = false
			break
		}
	}
	wsWriter.WriteWithSuccess([]byte(rId), testsPassed)
}

func (s *Server) handleExecuteShell(ctx context.Context, buf []byte, sessionKey string, c *websocket.Conn) {
	var shellRequest model.ShellRequest
	if err := request.ParseAndValidateRequest(buf, &shellRequest, c); err != nil {
		return
	}

	if shellRequest.Stdin == "" {
		wswriter.NewWriter(c, wswriter.WriteError).Write([]byte("Empty stdin"))
		return
	}

	err := shell.ShellExecute(ctx, shell.ShellExecuteParams{SessionKey: sessionKey, Stdin: shellRequest.Stdin, CodeRunner: s.CodeRunner})
	if err != nil {
		wswriter.NewWriter(c, wswriter.WriteError).Write([]byte(errorutil.ErrorWrap(err, "Execute/shell failed").Error()))
	}
}

func (s *Server) getDefaultTimeout(cmd string) int {
	cconfig := config.GetContainerConfig(cmd)
	if cconfig != nil && cconfig.Timeout != 0 {
		return cconfig.Timeout
	}
	return defaultTimeout
}
