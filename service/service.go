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
	GetFiles() []myFile
	UploadFile(uploadingFile io.Reader, fileName string) error
	CutFile(fileName string, dX int, dY int) error
	GetArchiveName(fileName string) (string, error)
}

type Service struct {
	Files   FileService
	Session SessionService
}

func NewService() Service {
	return Service{
		Files:   &fileManager{tempFiles: map[string]myFile{}},
		Session: &sessionManager{sessions: map[string]*Session{}},
	}
}
