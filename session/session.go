package session

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

var (
	sessions map[string]*Session
	mu       sync.RWMutex
	once     sync.Once
)

type Session struct {
	ContainerID string
	CmdID       string
	Con         io.ReadWriteCloser
	CancelFunc  context.CancelFunc
	Updated     time.Time
}

func init() {
	once.Do(func() {
		sessions = make(map[string]*Session, 0)
	})
}

func PutSession(key string, val *Session) *Session {
	mu.Lock()
	defer mu.Unlock()
	sessions[key] = val
	return val
}

func GetSession(key string) (*Session, error) {
	mu.RLock()
	defer mu.RUnlock()
	if session, ok := sessions[key]; ok {
		return session, nil
	}
	return nil, fmt.Errorf("no session available")
}

func GetSessions() map[string]*Session {
	return sessions
}

func DeleteSession(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(sessions, key)
}

func StopSession(key string) {
	mu.Lock()
	defer mu.Unlock()

	sess, ok := sessions[key]
	if !ok || sess == nil {
		return
	}

	if sess.Con != nil {
		_ = sess.Con.Close()
	}

	if sess.CancelFunc != nil {
		sess.CancelFunc()
	}
}
