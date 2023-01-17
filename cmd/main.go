package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type uploadingFileManager struct {
	ts        *template.Template
	tempFiles []*os.File
}

// func (h uploadingFileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}

// 	switch r.Method {
// 	case http.MethodGet:
// 		h.uploadFileSetup(w, r)
// 	case http.MethodPost:
// 		h.uploadFile(w, r)
// 	}
// }

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
	switch contentType {
	case "image/jpeg", "image/png":
		localFile, err = os.Create(fmt.Sprintf("temp/%s", fileHeader.Filename))
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
	localFile.Write(fileBytes)
	h.tempFiles = append(h.tempFiles, localFile)

	w.WriteHeader(http.StatusBadRequest)
	log.Print("file succsesfully uploaded")
	fmt.Fprint(w, "file succsesfully uploaded")

}

func newUploadingFileManager(template *template.Template) uploadingFileManager {
	return uploadingFileManager{
		ts:        template,
		tempFiles: []*os.File{},
	}
}

func main() {
	ts, err := template.New("home.html").Funcs(template.FuncMap{
		"Basing": filepath.Base,
	}).ParseFiles("./static/templates/home.html")
	if err != nil {
		log.Println(err.Error())
		return
	}

	//fmt.Printf("ts.Tree: %v\n", ts.Tree)
	uploadingFileHandler := newUploadingFileManager(ts)

	http.HandleFunc("/", uploadingFileHandler.uploadFileSetup)
	http.HandleFunc("/upload", uploadingFileHandler.uploadFile)
	http.ListenAndServe(":8080", nil)

}
