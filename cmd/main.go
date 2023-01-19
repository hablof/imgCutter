package main

import (
	"bytes"
	"fmt"
	"html/template"
	"imgCutter/internal"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"
)

type myFile struct {
	Name    string
	Archive string
}

type uploadingFileManager struct {
	ts        *template.Template
	tempFiles []myFile
}

func (h *uploadingFileManager) cutFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("err parsing form: %v", err)
		return
	}
	fileName := r.PostForm.Get("fileName")
	dX, err := strconv.Atoi(r.PostForm.Get("dX"))
	if err != nil {
		log.Printf("error parsing dX: %v", err)
		return
	}
	dY, err := strconv.Atoi(r.PostForm.Get("dY"))
	if err != nil {
		log.Printf("error parsing dY: %v", err)
		return
	}
	fmt.Printf("fileName: %v\n, dX: %v\n, dY: %v\n", fileName, dX, dY)

	archiveName, err2 := internal.ProcessImage(fileName, dX, dY)
	if err2 != nil {
		log.Printf("error processing img: %v", err2)
		return
	}

	for i, elem := range h.tempFiles {
		if strings.HasSuffix(elem.Name, fileName) {
			h.tempFiles[i].Archive = archiveName
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully cut", fileName)
	h.ts.ExecuteTemplate(w, "cutGood.html", fileName)
}

func (h *uploadingFileManager) uploadFileSetup(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if err := h.ts.ExecuteTemplate(&buf, "home.html", h.tempFiles); err != nil {
		log.Println(err.Error())
		log.Print(buf.String())
		w.WriteHeader(500)
		fmt.Fprint(w, "Internal Server Error")
		return
	}
	w.Write(buf.Bytes())
}

func (h *uploadingFileManager) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	uploadingFile, fileHeader, err := r.FormFile("uploadingFile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer uploadingFile.Close()

	var localFile *os.File
	contentType := fileHeader.Header.Get("content-type")
	fileName := fileHeader.Filename
	switch contentType {
	case "image/jpeg", "image/png":
		localFile, err = os.Create(fmt.Sprintf("temp/%s", fileName))
		if err != nil {
			log.Printf("error uploading file: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "error uploading file")
			return
		}
		defer localFile.Close()
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "file must be ДЖИПЕГ (or .png)")
		return
	}

	fileBytes, err := io.ReadAll(uploadingFile)
	if err != nil {
		log.Printf("error reading uploading file: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "error uploading file")
		return
	}
	n, err := localFile.Write(fileBytes)
	if err != nil {
		log.Printf("error writing bytes to localfile: %s", err)
		return
	}
	log.Printf("written %d bytes", n)
	if err := localFile.Sync(); err != nil {
		log.Printf("error sync written file: %s", err)
	}

	if newOffset, err := localFile.Seek(0, 0); err != nil {
		log.Printf("error seeking uploaded file: %s", err)
		return
	} else {
		log.Printf("new offset: %d", newOffset)
	}

	h.tempFiles = append(h.tempFiles, myFile{Name: localFile.Name()})
	log.Printf("uploadFile(w, r): localFile.Name(): %v\n", localFile.Name())
	log.Printf("h.tempFiles: %v\n", h.tempFiles)

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully uploaded", fileName)
	h.ts.ExecuteTemplate(w, "uploadGood.html", fileName)
}

func newUploadingFileManager(template *template.Template) uploadingFileManager {
	return uploadingFileManager{
		ts:        template,
		tempFiles: []myFile{},
	}
}

func main() {
	ts, err := template.New("home.html").Funcs(template.FuncMap{
		"base": filepath.Base,
	}).ParseGlob("./static/templates/*.html")
	if err != nil {
		log.Println(err.Error())
		return
	}

	fileManager := newUploadingFileManager(ts)

	http.HandleFunc("/", fileManager.uploadFileSetup)
	http.HandleFunc("/upload", fileManager.uploadFile)
	http.HandleFunc("/cut", fileManager.cutFile)
	http.ListenAndServe(":8080", nil)

}
