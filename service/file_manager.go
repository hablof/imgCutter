package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"imgCutter/imgProcessing"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrFS = errors.New("filesystem error")
)

type myFile struct {

	// Full-OriginalFile like path/OriginalFile.ext
	OriginalFile string // export to templates

	// Full-Name like path/Name.ext
	Archive string // export to templates

	uploaded time.Time
}

// key type is Full-Name like path/Name.ext
type tempFiles map[string]myFile

func (tf tempFiles) deleteFile(fileName string) error {
	f, ok := tf[fileName]
	if !ok {
		return errors.New("on such file")
	}
	if err := deleteFileIfExist(f.OriginalFile); err != nil {
		return err
	}
	if err := deleteFileIfExist(f.Archive); err != nil {
		return err
	}
	delete(tf, fileName)
	return nil
}

type Session struct {
	id               uuid.UUID
	files            tempFiles
	terminatnonTimer time.Timer
	terminateChannel chan (struct{})
	resetChannel     chan (struct{})
}

func (s *Session) String() string {
	return s.id.String()
}

type fileManager struct {
	sessions map[string]*Session
}

func (fm *fileManager) GetFiles(sessionID string) []myFile {
	output := make([]myFile, 0, len(fm.sessions[sessionID].files))
	for _, f := range fm.sessions[sessionID].files {
		output = append(output, f)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].uploaded.After(output[j].uploaded) })
	return output
}

func (fm *fileManager) CutFile(sessionID string, fileName string, dx int, dy int) error {
	// открываем изображение
	img, format, err := imgProcessing.OpenImage(fileName)
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
	images, err := imgProcessing.CutImage(img, dx, dy)
	if err != nil {
		log.Printf("error on cut img: %v", err)
		return err
	}

	// пакуем в архив
	if err := imgProcessing.PackImages(zipWriter, images, filepath.Base(archiveName)); err != nil {
		log.Printf("error on create archive file: %v", err)
		return err
	}

	// записываем путь архива в myFile
	if err := fm.setArchivePath(sessionID, fileName, archive.Name()); err != nil {
		log.Printf("error on set archive path: %v", err)
		return err
	}
	return nil
}

func (fm *fileManager) UploadFile(sessionID string, uploadingFile io.Reader, fileName string) error {
	var localFile *os.File

	if err := createDirIfNotExist(fmt.Sprintf("temp/%s", sessionID)); err != nil {
		log.Println(err)
		return ErrFS
	}

	localFile, err := os.Create(fmt.Sprintf("temp/%s/%s", sessionID, fileName))
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

	fm.sessions[sessionID].files[localFile.Name()] = myFile{
		OriginalFile: localFile.Name(),
		Archive:      "",
		uploaded:     time.Now(),
	}
	log.Printf("uploaded file: %v\n", localFile.Name())
	return nil
}

func (fm *fileManager) GetArchiveName(sessionID string, fileName string) (string, error) {
	f, ok := fm.sessions[sessionID].files[fileName]
	if !ok {
		return "", errors.New("on such file")
	}

	if err := checkFileExist(f.Archive); err != nil {
		return "", err
	}

	return f.Archive, nil
}

func (fm *fileManager) DeleteFile(sessionID string, fileName string) error {

	if err := fm.sessions[sessionID].files.deleteFile(fileName); err != nil {
		return err
	}

	return nil
}

func (fm *fileManager) setArchivePath(sessionID string, targetFileName string, archiveName string) error {
	f, ok := fm.sessions[sessionID].files[targetFileName]
	if !ok {
		return errors.New("on such file")
	}
	f.Archive = archiveName
	fm.sessions[sessionID].files[targetFileName] = f

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
