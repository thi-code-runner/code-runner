package session

import (
	"fmt"
	"net"
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
	Con         net.Conn
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
