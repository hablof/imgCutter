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
	NewSessionFiles(sessionID string)
}

type Service struct {
	Files   FileService
	Session SessionService
}

func NewService() Service {
	return Service{
		Files:   &fileManager{sessionFiles: map[string]tempFiles{}},
		Session: &sessionManager{sessions: map[string]*Session{}},
	}
}
