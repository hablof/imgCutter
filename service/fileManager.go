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
	"strings"
	"time"
)

var (
	ErrFS = errors.New("filesystem error")
)

type fileManager struct {
	tempFiles map[string]myFile
}

func (fm *fileManager) setArchivePath(targetFileName string, archiveName string) error {
	f, ok := fm.tempFiles[targetFileName]
	if !ok {
		return errors.New("on such file")
	}
	f.Archive = archiveName
	fm.tempFiles[targetFileName] = f

	return nil
}

func (fm *fileManager) GetFiles() []myFile {
	output := make([]myFile, 0, len(fm.tempFiles))
	for _, f := range fm.tempFiles {
		output = append(output, f)
	}
	return output
}

func (fm *fileManager) CutFile(fileName string, dx int, dy int) error {
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

	fm.setArchivePath(fileName, archiveName)
	return nil
}

func (fm *fileManager) UploadFile(uploadingFile io.Reader, fileName string) error {
	var localFile *os.File

	if err := createDirIfNotExist("temp"); err != nil {
		log.Println(err)
		return ErrFS
	}
	localFile, err := os.Create(fmt.Sprintf("temp/%s", fileName))
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

	fm.tempFiles[localFile.Name()] = myFile{
		Name:     localFile.Name(),
		Archive:  "",
		Uploaded: time.Now(),
	}
	log.Printf("uploaded file: %v\n", localFile.Name())
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
