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
	"html/template"
	"io"
	"log"
	"os"
	strings "strings"
	"sync"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerService interface {
	RunCommand(context.Context, string, container.RunCommandParams) (io.ReadWriteCloser, string, error)
	CreateAndStartContainer(context.Context, string, container.ContainerCreateParams) (string, error)
	PullImage(context.Context, string, io.Writer) error
	ContainerRemove(context.Context, string, container.RemoveCommandParams) error
	CopyToContainer(context.Context, string, []*model.SourceFile) error
	CopyFromContainer(context.Context, string, string) (string, error)
	GetReturnCode(context.Context, string) (int, error)
	GetContainers(context.Context) ([]string, error)
	CreateInteractiveShell(ctx context.Context, id string) (io.ReadWriteCloser, error)
	ExecuteCommand(ctx context.Context, con io.ReadWriteCloser, stdin string, withEndMarker bool) error
}
type Service struct {
	sync.Mutex
	sync.Once
	ContainerService ContainerService
	SchedulerService *scheduler.Scheduler
	containers       map[string]struct{}
}

func NewService(ctx context.Context, containerService ContainerService, schedulerService *scheduler.Scheduler) *Service {
	s := &Service{ContainerService: containerService, SchedulerService: schedulerService}
	s.Do(func() {
		var buf bytes.Buffer
		s.containers = make(map[string]struct{})
		for _, cc := range config.Conf.ContainerConfig {
			//pulling images of config file
			err := s.ContainerService.PullImage(ctx, cc.Image, &buf)
			if err != nil {
				log.Fatalf(errorutil.ErrorWrap(err, fmt.Sprintf("could not pull container image %s", cc.Image)).Error())
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
	containerConf := config.GetContainerConfig(cmdID)
	if containerConf == nil || containerConf.ID == "" {
		message := fmt.Errorf("no configuration found for %q", cmdID)
		return nil, "", message
	}
	sess, err := session.GetSession(sessionKey)
	if err == nil && cmdID == sess.CmdID {
		return containerConf, sess.ContainerID, nil
	}

	containerID, err := s.ContainerService.CreateAndStartContainer(ctx, containerConf.Image, container.ContainerCreateParams{Memory: containerConf.Memory, CPU: containerConf.CPU, ReadOnly: containerConf.ReadOnly, DiskSize: containerConf.DiskSize})
	if err != nil {
		return nil, "", err
	}

	func() {
		s.Lock()
		defer s.Unlock()
		s.containers[containerID] = struct{}{}
	}()

	session.PutSession(sessionKey, &session.Session{ContainerID: containerID, CmdID: cmdID, Updated: time.Now()})

	return containerConf, containerID, nil
}

func (s *Service) GetContainerConnection(ctx context.Context, sessionKey string, containerID string, writer wswriter.Writer, rId string) (io.ReadWriteCloser, error) {
	sess, err := session.GetSession(sessionKey)
	if err != nil || sess == nil {
		return nil, fmt.Errorf("could not retrieve session with key %q", sessionKey)
	} else {
		session.StopSession(sessionKey)
	}

	con, err := s.ContainerService.CreateInteractiveShell(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("could not create interactive shell: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	session.PutSession(sessionKey, &session.Session{
		ContainerID: containerID,
		Con:         con,
		CmdID:       sess.CmdID,
		Updated:     time.Now(),
		CancelFunc:  cancel,
	})

	// Pipe for separate reading of stdout and stderr
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	// stdcopy liest aus con und verteilt es auf die Writer
	go func() {
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		_, err := stdcopy.StdCopy(stdoutWriter, stderrWriter, con)
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			return
		}
		if err != nil && err != io.EOF {
			writer.WithType(wswriter.WriteError).Write([]byte("stdcopy error: " + err.Error()))
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := stdoutReader.Read(buf)
				if n > 0 {
					writer.WithType(wswriter.WriteOutput).Write(buf[:n])
				}
				if err != nil {
					if err != io.EOF {
						writer.WithType(wswriter.WriteError).Write([]byte("stdout read error: " + err.Error()))
					}
					return
				}
			}
		}
	}()

	go func() {
		hasErrors := false
		buf := make([]byte, 1024)
		lineBuf := bytes.Buffer{}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := stderrReader.Read(buf)
				if n > 0 {
					lineBuf.Write(buf[:n])

					for {
						line, err := lineBuf.ReadString('\n')
						if err != nil {
							// Line is not complete, stop and continue reading
							lineBuf.WriteString(line) // write back the potentially missing part
							break
						}

						trimmed := strings.TrimSpace(line)

						if trimmed == "" {
							continue
						}
						if trimmed == "__DONE__" {
							writer.WriteWithSuccess([]byte(rId), !hasErrors)
							hasErrors = false
							continue
						}
						if strings.Contains(strings.ToLower(trimmed), "error") {
							hasErrors = true
						}
						writer.WithType(wswriter.WriteError).Write([]byte(line))
					}
				}
				if err != nil {
					if err != io.EOF {
						writer.WithType(wswriter.WriteError).Write([]byte("stderr read error: " + err.Error()))
					}
					return
				}
			}
		}
	}()

	return con, nil
}

func (s *Service) Compile(ctx context.Context, containerID string, compilationCmd string, writer wswriter.Writer) error {
	if len(compilationCmd) > 0 {
		con, _, err := s.ContainerService.RunCommand(context.Background(), containerID, container.RunCommandParams{Cmd: compilationCmd})
		if err != nil {
			return err
		}
		defer con.Close()
		err = s.Copy(writer.WithType(wswriter.WriteError), con)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) Shutdown(ctx context.Context) {
	for id := range s.containers {
		_ = s.ContainerService.ContainerRemove(ctx, id, container.RemoveCommandParams{Force: true})
	}
}
func (s *Service) CopyWithTimeout(ctx context.Context) func(w io.Writer, r io.Reader) error {
	return func(w io.Writer, r io.Reader) error {
		var err error
		buf := make([]byte, 32*1024)
		ch := make(chan int)
		for {
			var n int
			var er error
			go func() {
				n, er = r.Read(buf)
				ch <- n
			}()
			select {
			case <-ctx.Done():
				return errorutil.TimeoutErr
			case <-ch:
				//noop
			}
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

type TransformParams struct {
	FileName string
}

func (s *Service) TransformCommand(cmd string, params TransformParams) (string, error) {
	tmpl, err := template.New("cmdTemplate").Funcs(template.FuncMap{
		"getSubstringUntil": func(txt string, delim string) string {
			return strings.Split(txt, delim)[0]
		},
	}).Parse(cmd)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
