package session

import "testing"

func TestPutSession(t *testing.T) {

	PutSession("key", &Session{})

	_, err := GetSession("key")
	if err != nil {
		t.Fail()
	}
}
func TestDeleteSession(t *testing.T) {

	PutSession("key", &Session{})
	DeleteSession("key")
	_, err := GetSession("key")
	if err == nil {
		t.Fail()
	}
}

func TestGetSessions(t *testing.T) {
	PutSession("key", &Session{})
	if len(GetSessions()) != 1 {
		t.Fail()
	}
}
