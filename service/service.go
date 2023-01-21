package service

import (
	"io"
	"time"
)

type myFile struct {
	Name     string
	Archive  string
	Uploaded time.Time
}

type FileService interface {
	GetFiles() []myFile
	UploadFile(uploadingFile io.Reader, fileName string) error
	CutFile(fileName string, dX int, dY int) error
}

type Service struct {
	FileService
}

func NewService() Service {
	return Service{
		FileService: &fileManager{
			tempFiles: map[string]myFile{},
		},
	}
}
