package service

import (
	"io"
)

type SessionService interface {
	Find(id string) (ses *Session, ok bool)
	New() *Session
	GetAll() []string //debug
}

type FileService interface {
	GetFiles(sessionID string) []myFile
	UploadFile(sessionID string, uploadingFile io.Reader, fileName string) error
	CutFile(sessionID string, fileName string, dX int, dY int) error
	GetArchiveName(sessionID string, fileName string) (string, error)
}

type Service struct {
	Files   FileService
	Session SessionService
}

func NewService() Service {
	mem := &fileManager{sessions: map[string]*Session{}}
	return Service{
		Files:   mem,
		Session: mem,
	}
}
