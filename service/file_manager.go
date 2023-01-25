package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"imgCutter/internal"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	ErrFS = errors.New("filesystem error")
)

type myFile struct {
	Name     string
	Archive  string
	Uploaded time.Time
}

type tempFiles map[string]myFile

// struct {
// 	tempFiles
// }

type fileManager struct {
	sessionFiles map[string]tempFiles
}

func (fm *fileManager) GetFiles(sessionID string) []myFile {
	output := make([]myFile, 0, len(fm.sessionFiles[sessionID]))
	for _, f := range fm.sessionFiles[sessionID] {
		output = append(output, f)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Uploaded.After(output[j].Uploaded) })
	return output
}

func (fm *fileManager) CutFile(sessionID string, fileName string, dx int, dy int) error {
	// открываем изображение
	img, format, err := internal.OpenImage(fileName)
	if err != nil {
		return err
	}
	log.Printf("Decoded format is: %s", format)

	// создаём архив
	archiveName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	archive, err := os.Create(fmt.Sprintf("%s.zip", archiveName))
	if err != nil {
		log.Printf("error on create archive file: %v", err)
		return err
	}

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	// режем изображение
	images, err := internal.CutImage(img, dx, dy)
	if err != nil {
		log.Printf("error on cut img: %v", err)
		return err
	}

	// пакуем в архив
	if err := internal.PackImages(zipWriter, images, filepath.Base(archiveName)); err != nil {
		log.Printf("error on create archive file: %v", err)
		return err
	}

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

	fm.sessionFiles[sessionID][localFile.Name()] = myFile{
		Name:     localFile.Name(),
		Archive:  "",
		Uploaded: time.Now(),
	}
	log.Printf("uploaded file: %v\n", localFile.Name())
	return nil
}

func (fm *fileManager) GetArchiveName(sessionID string, fileName string) (string, error) {
	f, ok := fm.sessionFiles[sessionID][fileName]
	if !ok {
		return "", errors.New("on such file")
	}

	if err := fm.checkFileExist(f.Archive); err != nil {
		return "", err
	}

	return f.Archive, nil
}

func (fm *fileManager) NewSessionFiles(sessionID string) {
	fm.sessionFiles[sessionID] = map[string]myFile{}
}

func (fm *fileManager) setArchivePath(sessionID string, targetFileName string, archiveName string) error {
	f, ok := fm.sessionFiles[sessionID][targetFileName]
	if !ok {
		return errors.New("on such file")
	}
	f.Archive = archiveName
	fm.sessionFiles[sessionID][targetFileName] = f

	return nil
}

func (fm *fileManager) checkFileExist(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return errors.New("on such file")
	}
	return nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
