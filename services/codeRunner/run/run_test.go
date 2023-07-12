package run

//import (
//	"bytes"
//	"code-runner/config"
//	"code-runner/model"
//	"code-runner/network/wswriter"
//	"code-runner/services/codeRunner"
//	"code-runner/services/container"
//	"code-runner/services/scheduler"
//	"code-runner/session"
//	"context"
//	"fmt"
//	"io"
//	"nhooyr.io/websocket"
//	"strings"
//	"testing"
//	"time"
//)
//
//var cr *codeRunner.Service = nil
//
//type ContainerServiceMock struct {}
//
//type conMock struct {
//	bytes.Buffer
//}
//
//func before() {
//	config.Conf = &config.Config{
//		ContainerConfig: []config.ContainerConfig{
//			{
//				ID:                     "test-id",
//				Image:                  "test-image",
//				ReserveContainerAmount: 2,
//			},
//			{
//				ID:    "test-id-no-copy",
//				Image: "test-image-no-copy",
//			},
//		},
//	}
//	cs := ContainerServiceMock{}
//	s := scheduler.NewScheduler(time.Millisecond)
//	service := codeRunner.NewService(context.Background(), &cs, s)
//	cr = service
//
//}
//func TestCRSavesContainerInMap(t *testing.T) {
//	before()
//	ctx := context.Background()
//	c := &websocket.Conn{}
//	wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
//	err := Run(ctx, "test-id", ExecuteParams{Writer: wsWriter})
//	if err != nil {
//		t.Error(err)
//	}
//	if _, ok := cr.containers["mock-container-id"]; !ok {
//		t.Fail()
//	}
//}
//func TestCRSavesReservedContainerInMap(t *testing.T) {
//	before()
//	ctx := context.Background()
//	c := &websocket.Conn{}
//	wsWriter := wswriter.NewWriter(c, wswriter.WriteOutput)
//	err := cr.Run(ctx, "test-id", ExecuteParams{Writer: wsWriter})
//	if err != nil {
//		t.Error(err)
//	}
//	if _, ok := cr.reservedContainers["test-image"]; !ok {
//		t.Fail()
//	}
//}
//func TestCRSavesContainerInSession(t *testing.T) {
//	before()
//	ctx := context.Background()
//	var wsBuf bytes.Buffer
//	wsWriter := wswriter.WSWriter{Con: &wsBuf, Type: wswriter.WriteOutput}
//	err := cr.Run(ctx, "test-id", ExecuteParams{Writer: &wsWriter})
//	if err != nil {
//		t.Error(err)
//	}
//	sessions := session.GetSessions()
//	var found bool
//	for _, s := range sessions {
//		if s.ContainerID == "mock-container-id" {
//			found = true
//			break
//		}
//	}
//	if !found {
//		t.Fail()
//	}
//}
//func TestCRNoCmdID(t *testing.T) {
//	before()
//	ctx := context.Background()
//	var wsBuf bytes.Buffer
//	wsWriter := wswriter.WSWriter{Con: &wsBuf, Type: wswriter.WriteOutput}
//	err := cr.Run(ctx, "nonexistent-cmdID", ExecuteParams{Writer: &wsWriter})
//	if err == nil {
//		t.Error(err)
//	}
//	if !strings.Contains(err.Error(), "no configuration found for \"nonexistent-cmdID\"") {
//		t.Fail()
//	}
//}
//func TestCRNoContainerID(t *testing.T) {
//	before()
//	ctx := context.Background()
//	var wsBuf bytes.Buffer
//	wsWriter := wswriter.WSWriter{Con: &wsBuf, Type: wswriter.WriteOutput}
//	err := cr.Run(ctx, "test-id-no-copy", ExecuteParams{Writer: &wsWriter})
//	if err == nil {
//		t.Error(err)
//	}
//	if !strings.Contains(err.Error(), "could not add files to sandbox environment") {
//		t.Fail()
//	}
//}
//func (m *conMock) Close() error { m.Reset(); return nil }
//func (cs *ContainerServiceMock) RunCommand(ctx context.Context, s string, params container.RunCommandParams) (io.ReadWriteCloser, string, error) {
//	if len(s) <= 0 {
//		return nil, "", fmt.Errorf("mock-error")
//	}
//	var buf bytes.Buffer
//	con := conMock{buf}
//	return &con, "", nil
//}
//func (cs *ContainerServiceMock) CreateAndStartContainer(ctx context.Context, s string) (string, error) {
//	if strings.Contains(s, "test-image-no-copy") {
//		return "", nil
//	}
//	return "mock-container-id", nil
//}
//func (cs *ContainerServiceMock) PullImage(context.Context, string, io.Writer) error { return nil }
//func (cs *ContainerServiceMock) ContainerRemove(context.Context, string, container.RemoveCommandParams) error {
//	return nil
//}
//func (cs *ContainerServiceMock) CopyToContainer(ctx context.Context, s string, files []*model.SourceFile) error {
//	if len(s) <= 0 {
//		return fmt.Errorf("mock-error")
//	}
//	return nil
//}
//func (cs *ContainerServiceMock) CopyResourcesToContainer(context.Context, string, []string) error {
//	return nil
//}
//func (cs *ContainerServiceMock) GetReturnCode(context.Context, string) (int, error) { return 0, nil }
//func (cs *ContainerServiceMock) GetContainers(context.Context) ([]string, error)    { return nil, nil }
