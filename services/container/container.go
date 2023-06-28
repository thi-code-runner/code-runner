package container

import (
	"code-runner/config"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"log"
	"os"
	"sync"
	"time"
)

var once sync.Once
var mu sync.Mutex

type Service struct {
	cli                *client.Client
	reservedContainers map[string][]string
	containers         map[string]struct{}
}

func (cs *Service) RunCommand(ctx context.Context, id string, params RunCommandParams) error {
	var containerConf config.ContainerConfig
	for _, c := range config.Conf.ContainerConfig {
		if params.CmdID == c.ID {
			containerConf = c
			break
		}
	}
	if _, ok := cs.containers[id]; !ok {
		relevantContainers := cs.reservedContainers[containerConf.Image]
		if len(relevantContainers) > 0 {
			mu.Lock()
			relevantContainer := relevantContainers[len(relevantContainers)-1]
			cs.containers[relevantContainer] = struct{}{}
			id = relevantContainer
			cs.reservedContainers[containerConf.Image] = cs.reservedContainers[containerConf.Image][:len(relevantContainers)-1]
			mu.Unlock()
		} else {
			var err error
			id, err = cs.createAndStartContainer(ctx, containerConf.Image)
			if err != nil {
				return err
			}
		}
	}

	if len(containerConf.CompilationCmd) > 0 {
		compilation, err := cs.cli.ContainerExecCreate(ctx, id, types.ExecConfig{AttachStderr: true, AttachStdout: true, Tty: true, WorkingDir: "/src", Cmd: []string{"sh", "-c", containerConf.CompilationCmd} /*[]string{"/bin/bash", "-c", params.Payload.Cmd}*/})
		if err != nil {
			return err
		}
		hijackedResponse, err := cs.cli.ContainerExecAttach(ctx, compilation.ID, types.ExecStartCheck{Tty: true, Detach: false})
		defer hijackedResponse.Close()
	}
	execution, err := cs.cli.ContainerExecCreate(ctx, id, types.ExecConfig{AttachStderr: true, AttachStdout: true, Tty: true, WorkingDir: "/src", Cmd: []string{"sh", "-c", containerConf.ExecutionCmd} /*[]string{"/bin/bash", "-c", params.Payload.Cmd}*/})
	if err != nil {
		return err
	}
	hijackedResponse, err := cs.cli.ContainerExecAttach(ctx, execution.ID, types.ExecStartCheck{Tty: true, Detach: false})
	defer hijackedResponse.Close()
	return nil
}
func NewService() *Service {
	var cs Service
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("could not connect to docker engine because of %q", err)
	}
	cs.cli = cli
	once.Do(func() {
		cs.reservedContainers = make(map[string][]string)
		cs.containers = make(map[string]struct{})
		for _, cc := range config.Conf.ContainerConfig {
			//pulling images of config file
			image, err := cli.ImagePull(context.Background(), cc.Image, types.ImagePullOptions{})
			os.Stdout.ReadFrom(image)
			err = image.Close()
			if err != nil {
				log.Printf("encountered error while pulling image %q: %q. \ncode-runner is starting anyway", cc.Image, err)
			}
			//reserving reservedContainers
			if cc.ReserveContainerAmount > 0 {
				for i := 0; i < cc.ReserveContainerAmount; i++ {
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Millisecond)
						defer cancel()
						id, _ := cs.createAndStartContainer(ctx, cc.Image)
						mu.Lock()
						cs.reservedContainers[cc.Image] = append(cs.reservedContainers[cc.Image], id)
						mu.Unlock()
					}()
				}
			}
		}
	})
	return &cs
}

func (cs *Service) createAndStartContainer(ctx context.Context, image string) (string, error) {
	containerName := fmt.Sprintf("code-runner-container-%s", uuid.New().String())
	resp, err := cs.cli.ContainerCreate(ctx, &container.Config{
		Image:      image,
		Cmd:        []string{"/bin/sh"},
		WorkingDir: "/src",
		Tty:        true,
	}, &container.HostConfig{NetworkMode: "none", AutoRemove: true}, nil, nil, containerName)
	err = cs.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (cs *Service) Shutdown(ctx context.Context) {
	for _, v := range cs.reservedContainers {
		for _, id := range v {
			_ = cs.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
		}
	}
	for id := range cs.containers {
		_ = cs.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
	}
}
