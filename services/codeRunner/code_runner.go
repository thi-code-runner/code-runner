package codeRunner

import (
	"bytes"
	"code-runner/config"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/services/container"
	"code-runner/session"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"nhooyr.io/websocket"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex
var once sync.Once

type ContainerService interface {
	RunCommand(context.Context, string, container.RunCommandParams) (net.Conn, error)
	CreateAndStartContainer(context.Context, string) (string, error)
	PullImage(context.Context, string, io.Writer) error
	ContainerRemove(context.Context, string, container.RemoveCommandParams) error
	CopyToContainer(context.Context, string, []*model.SourceFile) error
}
type Service struct {
	ContainerService   ContainerService
	reservedContainers map[string][]string
	containers         map[string]struct{}
}

func NewService(ctx context.Context, containerService ContainerService) *Service {
	s := &Service{ContainerService: containerService}
	once.Do(func() {
		var buf bytes.Buffer
		s.reservedContainers = make(map[string][]string)
		s.containers = make(map[string]struct{})
		for _, cc := range config.Conf.ContainerConfig {
			//pulling images of config file
			s.ContainerService.PullImage(ctx, cc.Image, &buf)
			//reserving reservedContainers
			if cc.ReserveContainerAmount > 0 {
				for i := 0; i < cc.ReserveContainerAmount; i++ {
					func() {
						ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
						defer cancel()
						id, _ := s.ContainerService.CreateAndStartContainer(ctx, cc.Image)
						mu.Lock()
						s.reservedContainers[cc.Image] = append(s.reservedContainers[cc.Image], id)
						mu.Unlock()
					}()
				}
			}
		}
		io.Copy(os.Stdout, &buf)
	})
	return s
}
func (s *Service) Execute(ctx context.Context, id string, params ExecuteParams) error {
	var containerConf config.ContainerConfig
	for _, c := range config.Conf.ContainerConfig {
		if id == c.ID {
			containerConf = c
			break
		}
	}
	var containerID string
	sess, err := session.GetSession(params.SessionKey)
	if err == nil {
		containerID = sess.ContainerID
	}
	containerID, err = s.getContainerID(ctx, containerID, containerConf)
	if err != nil {
		return err
	}
	sess = session.PutSession(params.SessionKey, &session.Session{ContainerID: containerID})
	err = s.ContainerService.CopyToContainer(ctx, containerID, params.Files)
	if len(containerConf.CompilationCmd) > 0 {
		if err != nil {
			return errorutil.ErrorWrap(err, "could not create writer")
		}
		con, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.CompilationCmd})
		defer con.Close()
		s.copy(ctx, params.Con, con)
		if err != nil {
			return errorutil.ErrorWrap(err, "compilation failed")
		}
	}
	if err != nil {
		return errorutil.ErrorWrap(err, "could not create writer")
	}
	con, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.ExecutionCmd})
	defer con.Close()
	sess.Con = con
	err = s.copy(ctx, params.Con, con)
	if err != nil {
		return errorutil.ErrorWrap(err, "execution failed")
	}
	return nil
}
func (s *Service) getContainerID(ctx context.Context, containerID string, containerConf config.ContainerConfig) (string, error) {
	if _, ok := s.containers[containerID]; !ok {
		relevantContainers := s.reservedContainers[containerConf.Image]
		if len(relevantContainers) > 0 {
			mu.Lock()
			relevantContainer := relevantContainers[len(relevantContainers)-1]
			s.containers[relevantContainer] = struct{}{}
			containerID = relevantContainer
			s.reservedContainers[containerConf.Image] = s.reservedContainers[containerConf.Image][:len(relevantContainers)-1]
			mu.Unlock()
		} else {
			var err error
			containerID, err = s.ContainerService.CreateAndStartContainer(ctx, containerConf.Image)
			if err != nil {
				return "", err
			}
		}
	}
	return containerID, nil
}

func (s *Service) SendStdIn(ctx context.Context, stdin string, sessionKey string) error {
	var err error
	sess, err := session.GetSession(sessionKey)
	if err != nil || sess.Con == nil {
		return fmt.Errorf("no running execution found to pass stdin to")
	}
	_, err = sess.Con.Write(append([]byte(stdin), '\n'))
	return err
}
func (s *Service) Shutdown(ctx context.Context) {
	for _, v := range s.reservedContainers {
		for _, id := range v {
			_ = s.ContainerService.ContainerRemove(ctx, id, container.RemoveCommandParams{Force: true})
		}
	}
	for id := range s.containers {
		_ = s.ContainerService.ContainerRemove(ctx, id, container.RemoveCommandParams{Force: true})
	}
}

func (s *Service) copy(ctx context.Context, w *websocket.Conn, r io.Reader) error {
	var err error
	buf := make([]byte, 32*1024)
	for {
		n, er := r.Read(buf)
		if n > 0 {
			resp := model.RunResponse{Output: string(buf[0:n])}
			respJson, err := json.Marshal(resp)
			if err != nil {
				return err
			}
			w.Write(ctx, 1, respJson)
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return err
}
