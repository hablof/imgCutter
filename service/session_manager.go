package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	SessionLifetime time.Duration = 1 * time.Minute
)

func (fm *fileManager) ResetTimer(sessionID string) {
	s, ok := fm.Find(sessionID)
	if !ok {
		return
	}
	log.Printf("%s has found", sessionID)
	// убиваем горутину
	s.resetChannel <- struct{}{}
	log.Printf("%s gorutine killed", sessionID)

	// безопасно ресетим не вызывая терминацию
	if !s.terminatnonTimer.Stop() {
		<-s.terminatnonTimer.C
	}
	s.terminatnonTimer.Reset(SessionLifetime)
	log.Printf("%s timer updated", sessionID)

	// опять запускаем горутину НАДЁЖНО
	var wg sync.WaitGroup
	wg.Add(1)
	go fm.terminationTrigger(s, &wg)
	wg.Wait()
	log.Printf("%s timer reseted", sessionID)
}

func (fm *fileManager) terminationTrigger(s *Session, wg *sync.WaitGroup) {
	wg.Done()
	log.Printf("%s trigger is set", s.String())
	for {
		select {
		case <-s.terminatnonTimer.C:
			log.Printf("%s timer triggered", s.String())
			go func() { s.terminateChannel <- struct{}{} }()
		case <-s.terminateChannel:
			// log.Println("terminate channel triggered")
			s.terminatnonTimer.Stop()
			// log.Println("timer stopped")
			fm.terminate(s.id.String())
			return
		case <-s.resetChannel:
			return
		}
	}
}

func (fm *fileManager) Find(id string) (ses *Session, ok bool) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}
	ses, ok = fm.sessions[id]
	return ses, ok
}

func (fm *fileManager) New() *Session {
	log.Printf("creating new session")
	var s Session
	for {
		s = Session{
			id:               uuid.New(),
			files:            map[string]myFile{},
			terminatnonTimer: time.Timer{},
			terminateChannel: make(chan struct{}),
			resetChannel:     make(chan struct{}),
		}
		if _, ok := fm.sessions[s.id.String()]; !ok {
			break
		}
	}
	s.terminatnonTimer = *time.NewTimer(SessionLifetime)
	fm.sessions[s.id.String()] = &s

	var wg sync.WaitGroup
	wg.Add(1)
	go fm.terminationTrigger(&s, &wg)
	wg.Wait()

	return &s
}

func (fm *fileManager) TerminateSession(sessionID string) error {
	s, ok := fm.Find(sessionID)
	if !ok {
		return errors.New("session not found")
	}
	s.terminateChannel <- struct{}{}
	return nil
}

func (fm *fileManager) terminate(sessionID string) error {
	s, ok := fm.Find(sessionID)
	if !ok {
		return errors.New("session not found")
	}
	log.Printf("terminating session %s...", s.String())
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
