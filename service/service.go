package service

import (
	"io"
	"sync"
)

type SessionService interface {
	Find(id string) (s *Session, ok bool)
	New() *Session
	GetAll() []string // debug
	TerminateSession(sessionID string) error
}

type FileService interface {
	GetFiles(s *Session) ([]myFile, error)
	UploadFile(s *Session, uploadingFile io.Reader, fileName string) error
	CutFile(s *Session, fileName string, dX int, dY int) error
	DeleteFile(s *Session, fileName string) error
	GetArchiveName(s *Session, fileName string) (string, error)
}

type Service struct {
	Files   FileService
	Session SessionService
}

func NewService() Service {
	mem := &fileManager{
		sessionsMapMutex: sync.Mutex{},
		sessions:         map[string]*Session{},
	}

	return Service{
		Files:   mem,
		Session: mem,
	}
}
