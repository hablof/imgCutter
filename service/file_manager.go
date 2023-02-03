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

var (
	ErrFileNotFound = errors.New("file not found")
	ErrFS           = errors.New("filesystem error")
	ErrNilSession   = errors.New("nil session")
)

type MyFile struct {
	// Full-OriginalFile like path/OriginalFile.ext
	OriginalFile string // export to templates

	// Full-Name like path/Name.ext
	Archive string // export to templates

	uploaded time.Time
}

// key type is Full-Name like path/Name.ext .
type tempFiles map[string]MyFile

func (tf tempFiles) deleteFile(fileName string) error {
	file, ok := tf[fileName]
	if !ok {
		return ErrFileNotFound
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
	fileMutex sync.Mutex // лочим на работу с мапой tempFiles и на всё из пакета "os" (Create, Open, Mkdir...)
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

func (fm *fileManager) GetFiles(s *Session) ([]MyFile, error) {
	if s == nil {
		return nil, ErrNilSession
	}

	output := make([]MyFile, 0, len(s.files))
	for _, f := range s.files {
		output = append(output, f)
	}

	sort.Slice(output, func(i, j int) bool { return output[i].uploaded.After(output[j].uploaded) })

	return output, nil
}

func (fm *fileManager) CutFile(s *Session, fileName string, dx int, dy int) error {
	if s == nil {
		return ErrNilSession
	}

	// дальше функция состоит почти полностью из работы с файлами
	s.fileMutex.Lock()
	defer s.fileMutex.Unlock()

	// открываем изображение
	img, format, err := imgprocessing.OpenImage(fileName)
	if err != nil {
		return fmt.Errorf("error processing image: %w", err)
	}

	log.Printf("Decoded format is: %s", format)

	// режем изображение
	images, err := imgprocessing.CutImage(img, dx, dy)
	if err != nil {
		e := fmt.Errorf("error on cut img: %w", err)
		log.Println(e)
		return e
	}

	// создаём архив
	// archiveName = path + name, без расширениея
	archiveName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	archive, err := os.Create(fmt.Sprintf("%s.zip", archiveName))

	if err != nil {
		e := fmt.Errorf("error on create archive file: %w", err)
		log.Println(e)

		return e
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	// пакуем в архив
	if err := imgprocessing.PackImages(zipWriter, images, filepath.Base(archiveName)); err != nil {
		e := fmt.Errorf("error on create archive file: %w", err)
		log.Println(e)
		return e
	}

	// записываем путь архива в myFile
	if err := fm.setArchivePath(s, fileName, archive.Name()); err != nil {
		e := fmt.Errorf("error on set archive path: %w", err)
		log.Println(e)
		return e
	}

	return nil
}

func (fm *fileManager) UploadFile(session *Session, uploadingFile io.Reader, fileName string) error {
	if session == nil {
		return ErrNilSession
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

	session.files[localFile.Name()] = MyFile{
		OriginalFile: localFile.Name(),
		Archive:      "",
		uploaded:     time.Now(),
	}

	log.Printf("uploaded file: %v\n", localFile.Name())

	return nil
}

func (fm *fileManager) GetArchiveName(session *Session, fileName string) (string, error) {
	if session == nil {
		return "", ErrNilSession
	}

	f, ok := session.files[fileName]
	if !ok {
		return "", ErrFileNotFound
	}

	if err := checkFileExist(f.Archive); err != nil {
		return "", err
	}

	return f.Archive, nil
}

func (fm *fileManager) DeleteFile(session *Session, fileName string) error {
	if session == nil {
		return ErrNilSession
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
		return ErrNilSession
	}

	file, ok := s.files[targetFileName]
	if !ok {
		return ErrFileNotFound
	}

	file.Archive = archiveName
	s.files[targetFileName] = file

	return nil
}

func checkFileExist(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	return nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to mkdir: %w", err)
		}
	}

	return nil
}

func deleteFileIfExist(fileName string) error {
	// err может быть либо "nil" либо "ErrFileNotFound"
	if err := checkFileExist(fileName); errors.Is(err, ErrFileNotFound) {
		return nil
	}

	// удвляем те файлы, которые существуют
	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("unable to remove: %w", err)
	}

	return nil
}
