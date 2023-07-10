package codeRunner

import (
	"bytes"
	"code-runner/config"
	"code-runner/model"
	"code-runner/services/container"
	"code-runner/services/scheduler"
	"code-runner/session"
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
	CopyResourcesToContainer(context.Context, string, []string) error
	GetReturnCode(context.Context, string) (int, error)
	GetContainers(context.Context) ([]string, error)
}
type Service struct {
	sync.Mutex
	sync.Once
	ContainerService   ContainerService
	SchedulerService   *scheduler.Scheduler
	reservedContainers map[string][]string
	containers         map[string]struct{}
}

func NewService(ctx context.Context, containerService ContainerService, schedulerService *scheduler.Scheduler) *Service {
	s := &Service{ContainerService: containerService, SchedulerService: schedulerService}
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
		s.SchedulerService.AddJob(&scheduler.Job{D: 30 * time.Second, Apply: func() {
			actualContainers, _ := s.ContainerService.GetContainers(ctx)
			actualContainerMap := make(map[string]struct{})
			for _, ac := range actualContainers {
				actualContainerMap[ac] = struct{}{}
			}
			for c, _ := range s.containers {
				if _, ok := actualContainerMap[c]; !ok {
					s.ContainerService.ContainerRemove(ctx, c, container.RemoveCommandParams{Force: true})
				}
			}
		}})
		s.SchedulerService.AddJob(&scheduler.Job{D: time.Minute, Apply: func() {
			for k, v := range session.GetSessions() {
				if v.Updated.Add(90 * time.Minute).Before(time.Now()) {
					err := s.ContainerService.ContainerRemove(ctx, v.ContainerID, container.RemoveCommandParams{Force: true})
					if err == nil {
						delete(s.containers, v.ContainerID)
						session.DeleteSession(k)
					}
				}
			}
		}})
		s.SchedulerService.Run(ctx)
	})
	return s
}
func (s *Service) getContainer(ctx context.Context, cmdID string, sessionKey string) (*config.ContainerConfig, string, error) {
	var containerConf config.ContainerConfig
	for _, c := range config.Conf.ContainerConfig {
		if cmdID == c.ID {
			containerConf = c
			break
		}
	}
	var containerID string
	sess, err := session.GetSession(sessionKey)
	if err == nil && cmdID == sess.CmdID {
		containerID = sess.ContainerID
	}
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
				return nil, "", err
			}
			func() {
				s.Lock()
				defer s.Unlock()
				s.containers[containerID] = struct{}{}
			}()
		}
		err := s.ContainerService.CopyResourcesToContainer(ctx, containerID, containerConf.Add)
		if err != nil {
			return nil, "", err
		}
	}
	sess = session.PutSession(sessionKey, &session.Session{ContainerID: containerID, CmdID: containerConf.ID, Updated: time.Now()})
	return &containerConf, containerID, nil
}
