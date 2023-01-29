package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
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
			files:     map[string]myFile{},
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

func (fm *fileManager) TerminateSession(sessionID string) error {
	_, ok := fm.Find(sessionID)
	if !ok {
		return errors.New("session not found")
	}

	if err := os.RemoveAll(fmt.Sprintf("temp/%s", sessionID)); err != nil {
		log.Printf("unable to remove session files: %v", err)
		return err
	}

	fm.sessionsMapMutex.Lock()
	defer fm.sessionsMapMutex.Unlock()

	delete(fm.sessions, sessionID)

	return nil
}

func (fm *fileManager) GetAll() []string {
	out := make([]string, 0, len(fm.sessions))
	for _, v := range fm.sessions {
		out = append(out, v.String())
	}

	return out
}
