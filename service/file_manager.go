package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"imgcutter/imgprocessing"

	"github.com/google/uuid"
)

var ErrFS = errors.New("filesystem error")

type myFile struct {
	// Full-OriginalFile like path/OriginalFile.ext
	OriginalFile string // export to templates

	// Full-Name like path/Name.ext
	Archive string // export to templates

	uploaded time.Time
}

// key type is Full-Name like path/Name.ext .
type tempFiles map[string]myFile

func (tf tempFiles) deleteFile(fileName string) error {
	file, ok := tf[fileName]
	if !ok {
		return errors.New("on such file")
	}

	if err := deleteFileIfExist(file.OriginalFile); err != nil {
		return err
	}

	if err := deleteFileIfExist(file.Archive); err != nil {
		return err
	}

	delete(tf, fileName)

	return nil
}

type Session struct {
	id        uuid.UUID
	fileMutex sync.Mutex // лочим на работу с мапой tempFiles и на всё, что из пакета "os" (Create, Open, Mkdir...)
	files     tempFiles
}

// returns string presintation of session's id.
func (s *Session) String() string {
	return s.id.String()
}

type fileManager struct {
	sessionsMapMutex sync.Mutex
	sessions         map[string]*Session
}

func (fm *fileManager) GetFiles(s *Session) ([]myFile, error) {
	if s == nil {
		return nil, errors.New("nil session")
	}

	output := make([]myFile, 0, len(s.files))
	for _, f := range s.files {
		output = append(output, f)
	}

	sort.Slice(output, func(i, j int) bool { return output[i].uploaded.After(output[j].uploaded) })

	return output, nil
}

func (fm *fileManager) CutFile(s *Session, fileName string, dx int, dy int) error {
	if s == nil {
		return errors.New("nil session")
	}

	// дальше функция состоит почти полностью из работы с файлами
	s.fileMutex.Lock()
	defer s.fileMutex.Unlock()

	// открываем изображение
	img, format, err := imgprocessing.OpenImage(fileName)
	if err != nil {
		return err
	}

	log.Printf("Decoded format is: %s", format)

	// создаём архив
	// archiveName = path + name
	archiveName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	archive, err := os.Create(fmt.Sprintf("%s.zip", archiveName))

	if err != nil {
		log.Printf("error on create archive file: %v", err)
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	// режем изображение
	images, err := imgprocessing.CutImage(img, dx, dy)
	if err != nil {
		log.Printf("error on cut img: %v", err)
		return err
	}

	// пакуем в архив
	if err := imgprocessing.PackImages(zipWriter, images, filepath.Base(archiveName)); err != nil {
		log.Printf("error on create archive file: %v", err)
		return err
	}

	// записываем путь архива в myFile
	if err := fm.setArchivePath(s, fileName, archive.Name()); err != nil {
		log.Printf("error on set archive path: %v", err)
		return err
	}

	return nil
}

func (fm *fileManager) UploadFile(session *Session, uploadingFile io.Reader, fileName string) error {
	if session == nil {
		return errors.New("nil session")
	}

	session.fileMutex.Lock()
	defer session.fileMutex.Unlock()

	var localFile *os.File

	if err := createDirIfNotExist(fmt.Sprintf("temp/%s", session.String())); err != nil {
		log.Println(err)
		return ErrFS
	}

	localFile, err := os.Create(fmt.Sprintf("temp/%s/%s", session.String(), fileName))
	if err != nil {
		log.Printf("error creating file: %s", err)
		return ErrFS
	}
	defer localFile.Close()

	fileBytes, err := io.ReadAll(uploadingFile)
	if err != nil {
		log.Printf("error reading uploading file: %s", err)
		return ErrFS
	}

	n, err := localFile.Write(fileBytes)
	if err != nil {
		log.Printf("error writing bytes to localfile: %s", err)
		return ErrFS
	}

	log.Printf("written %d bytes", n)

	if err := localFile.Sync(); err != nil {
		log.Printf("error sync written file: %s", err)
		return ErrFS
	}

	if _, err := localFile.Seek(0, 0); err != nil {
		log.Printf("error seeking uploaded file: %s", err)
		return ErrFS
	}

	session.files[localFile.Name()] = myFile{
		OriginalFile: localFile.Name(),
		Archive:      "",
		uploaded:     time.Now(),
	}

	log.Printf("uploaded file: %v\n", localFile.Name())

	return nil
}

func (fm *fileManager) GetArchiveName(s *Session, fileName string) (string, error) {
	if s == nil {
		return "", errors.New("nil session")
	}

	f, ok := s.files[fileName]
	if !ok {
		return "", errors.New("on such file")
	}

	if err := checkFileExist(f.Archive); err != nil {
		return "", err
	}

	return f.Archive, nil
}

func (fm *fileManager) DeleteFile(session *Session, fileName string) error {
	if session == nil {
		return errors.New("nil session")
	}

	session.fileMutex.Lock()
	defer session.fileMutex.Unlock()

	if err := session.files.deleteFile(fileName); err != nil {
		return err
	}

	return nil
}

func (fm *fileManager) setArchivePath(s *Session, targetFileName string, archiveName string) error {
	if s == nil {
		return errors.New("nil session")
	}

	file, ok := s.files[targetFileName]
	if !ok {
		return errors.New("on such file")
	}

	file.Archive = archiveName
	s.files[targetFileName] = file

	return nil
}

func checkFileExist(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return errors.New("on such file")
	}

	return nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteFileIfExist(fileName string) error {
	if err := checkFileExist(fileName); err != nil { // err может быть либо "nil" либо "errors.New("on such file")"
		return nil
	}

	// удвляем те файлы, которые существуют
	if err := os.Remove(fileName); err != nil {
		return err
	}

	return nil
}
