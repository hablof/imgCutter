package router

import (
	"bytes"
	"fmt"
	"html/template"
	"imgCutter/service"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type fileHandler struct {
	ts      *template.Template
	service service.Service
}

func (h *fileHandler) CutFile(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("fileName: %v, dX: %v, dY: %v", fileName, dX, dY)

	if err := h.service.CutFile(fileName, dX, dY); err != nil {
		log.Printf("error processing img: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully cut", fileName)
	h.ts.ExecuteTemplate(w, "cutGood.html", fileName)
}

func (h *fileHandler) MainPage(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if err := h.ts.ExecuteTemplate(&buf, "home.html", h.service.GetFiles()); err != nil {
		log.Println(err.Error())
		log.Print(buf.String())
		w.WriteHeader(500)
		fmt.Fprint(w, "Internal Server Error")
		return
	}
	w.Write(buf.Bytes())
}

func (h *fileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("err parsing form: %v", err)
		return
	}
	fileName := r.PostForm.Get("fileName")
	log.Printf("downloading archive of: %v", fileName)

	archiveName, err := h.service.GetArchiveName(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filepath.Base(archiveName)))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, archiveName)
}

func (h *fileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	uploadingFile, fileHeader, err := r.FormFile("uploadingFile")
	if err != nil {
		log.Println(err)
		return
	}
	defer uploadingFile.Close()
	contentType := fileHeader.Header.Get("content-type")
	fileName := fileHeader.Filename

	if !(contentType == "image/jpeg" || contentType == "image/png") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "file must be ДЖИПЕГ (or .png)")
		return
	}

	if err := h.service.UploadFile(uploadingFile, fileName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully uploaded", fileName)
	h.ts.ExecuteTemplate(w, "uploadGood.html", fileName)
}

func newFileHandler(template *template.Template, service service.Service) *fileHandler {
	return &fileHandler{
		ts:      template,
		service: service,
	}
}
