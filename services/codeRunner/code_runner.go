package codeRunner

import (
	"bytes"
	"code-runner/config"
	"code-runner/model"
	"code-runner/services/container"
	"context"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex
var once sync.Once

type ContainerService interface {
	RunCommand(context.Context, string, container.RunCommandParams) (net.Conn, string, error)
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

func (s *Service) copy(w io.Writer, r io.Reader) error {
	var err error
	buf := make([]byte, 32*1024)
	for {
		n, er := r.Read(buf)
		if n > 0 {
			w.Write(buf[0:n])
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
