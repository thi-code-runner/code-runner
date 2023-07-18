package codeRunner

import (
	"bytes"
	"code-runner/config"
	errorutil "code-runner/error_util"
	"code-runner/model"
	"code-runner/network/wswriter"
	"code-runner/services/container"
	"code-runner/services/scheduler"
	"code-runner/session"
	"context"
	"fmt"
	"io"
	"log"
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
	RunCommand(context.Context, string, container.RunCommandParams) (io.ReadWriteCloser, string, error)
	CreateAndStartContainer(context.Context, string) (string, error)
	PullImage(context.Context, string, io.Writer) error
	ContainerRemove(context.Context, string, container.RemoveCommandParams) error
	CopyToContainer(context.Context, string, []*model.SourceFile) error
	CopyFromContainer(context.Context, string, string) (string, error)
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
			err := s.ContainerService.PullImage(ctx, cc.Image, &buf)
			if err != nil {
				log.Fatalf(errorutil.ErrorWrap(err, fmt.Sprintf("could not pull container image %s", cc.Image)).Error())
			}
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
		s.SchedulerService.AddJob(&scheduler.Job{D: time.Duration(config.Conf.CacheCleanupIntervalS) * time.Second, Apply: func() {
			//Cleans up containers present in code-runner but not actually running on host system
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
		s.SchedulerService.AddJob(&scheduler.Job{D: time.Duration(config.Conf.HostCleanupIntervalS) * time.Second, Apply: func() {
			//clean up sessions and associated containers after a certain time of no usage
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
func (s *Service) GetContainer(ctx context.Context, cmdID string, sessionKey string) (*config.ContainerConfig, string, error) {
	var containerConf config.ContainerConfig
	for _, c := range config.Conf.ContainerConfig {
		if cmdID == c.ID {
			containerConf = c
			break
		}
	}
	if containerConf.ID == "" {
		message := fmt.Errorf("no configuration found for %q", cmdID)
		return nil, "", message
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
func (s *Service) Compile(ctx context.Context, containerID string, containerConf *config.ContainerConfig, writer wswriter.Writer) error {
	if len(containerConf.CompilationCmd) > 0 {
		con, _, err := s.ContainerService.RunCommand(ctx, containerID, container.RunCommandParams{Cmd: containerConf.CompilationCmd})
		if err != nil {
			return err
		}
		defer con.Close()
		err = s.Copy(writer.WithType(wswriter.WriteOutput), con)
		if err != nil {
			return err
		}
	}
	return nil
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
func (s *Service) Copy(w io.Writer, r io.Reader) error {
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
