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

var (
	mu               sync.Mutex
	once             sync.Once
	pullImageTimeout = 10000 * time.Millisecond
)

type ContainerService interface {
	RunCommand(context.Context, string, container.RunCommandParams) (net.Conn, string, error)
	CreateAndStartContainer(context.Context, string) (string, error)
	PullImage(context.Context, string, io.Writer) error
	ContainerRemove(context.Context, string, container.RemoveCommandParams) error
	CopyToContainer(context.Context, string, []*model.SourceFile) error
	GetReturnCode(context.Context, string) (int, error)
}
type Service struct {
	sync.Mutex
	sync.Once
	ContainerService   ContainerService
	reservedContainers map[string][]string
	containers         map[string]struct{}
}

func NewService(ctx context.Context, containerService ContainerService) *Service {
	s := &Service{ContainerService: containerService}
	s.Do(func() {
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
						ctx, cancel := context.WithTimeout(ctx, pullImageTimeout)
						defer cancel()
						id, _ := s.ContainerService.CreateAndStartContainer(ctx, cc.Image)
						s.Lock()
						defer s.Unlock()
						s.reservedContainers[cc.Image] = append(s.reservedContainers[cc.Image], id)
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
			func() {
				s.Lock()
				defer s.Unlock()
				relevantContainer := relevantContainers[len(relevantContainers)-1]
				s.containers[relevantContainer] = struct{}{}
				containerID = relevantContainer
				s.reservedContainers[containerConf.Image] = s.reservedContainers[containerConf.Image][:len(relevantContainers)-1]
			}()
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
