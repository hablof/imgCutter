package service

import (
	"io"
)

type FileService interface {
	GetFiles() []myFile
	UploadFile(uploadingFile io.Reader, fileName string) error
	CutFile(fileName string, dX int, dY int) error
	GetArchiveName(fileName string) (string, error)
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
