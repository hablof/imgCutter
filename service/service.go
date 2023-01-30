package service

import (
	"io"
	"sync"
)

//go:generate mockgen -source=service.go -destination=mock_service.go -package=service

type SessionService interface {
	Find(id string) (s *Session, ok bool)
	New() *Session
	// GetAll() []string // debug
	RemoveAll() error
	TerminateSession(session *Session) error
}

type FileService interface {
	GetFiles(s *Session) ([]MyFile, error)
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
