package service

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

func (fm *fileManager) Find(id string) (ses *Session, ok bool) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}

	ses, ok = fm.sessions[id]

	return ses, ok
}

func (fm *fileManager) New() *Session {
	var session Session

	for {
		session = Session{
			id:        uuid.New(),
			fileMutex: sync.Mutex{},
			files:     map[string]MyFile{},
		}
		if _, ok := fm.sessions[session.id.String()]; !ok {
			break
		}
	}

	fm.sessionsMapMutex.Lock()
	fm.sessions[session.id.String()] = &session
	fm.sessionsMapMutex.Unlock()

	return &session
}

func (fm *fileManager) TerminateSession(session *Session) error {
	if session == nil {
		return ErrNilSession
	}

	if err := os.RemoveAll(fmt.Sprintf("temp/%s", session.String())); err != nil {
		return fmt.Errorf("unable to remove session files: %w", err)
	}

	fm.sessionsMapMutex.Lock()
	defer fm.sessionsMapMutex.Unlock()

	delete(fm.sessions, session.String())

	return nil
}

func (fm *fileManager) RemoveAll() error {
	if err := os.RemoveAll("temp"); err != nil {
		return fmt.Errorf("unable to remove temp files: %w", err)
	}

	return nil
}

func (fm *fileManager) GetAll() []string {
	out := make([]string, 0, len(fm.sessions))
	for _, v := range fm.sessions {
		out = append(out, v.String())
	}

	return out
}
