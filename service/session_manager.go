package service

import (
	"errors"
	"fmt"
	"log"
	"os"

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
	var ses Session
	for {
		ses = Session{
			id:    uuid.New(),
			files: map[string]myFile{},
		}
		if _, ok := fm.sessions[ses.id.String()]; !ok {
			break
		}
	}
	fm.sessions[ses.id.String()] = &ses
	return &ses
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
