package router

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

func (h *Handler) CutFile(w http.ResponseWriter, r *http.Request) {
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

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := h.service.Files.CutFile(sessionID, fileName, dX, dY); err != nil {
		log.Printf("error processing img: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully cut", fileName)
	h.ts.ExecuteTemplate(w, "cutGood.html", fileName)
}

func (h *Handler) MainPage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
	}

	buf := bytes.Buffer{}
	if err := h.ts.ExecuteTemplate(&buf, "home.html", h.service.Files.GetFiles(sessionID)); err != nil {
		log.Println(err.Error())
		log.Print(buf.String())
		w.WriteHeader(500)
		fmt.Fprint(w, "Internal Server Error")
		return
	}
	w.Write(buf.Bytes())
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
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

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
	}

	archiveName, err := h.service.Files.GetArchiveName(sessionID, fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filepath.Base(archiveName)))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, archiveName)
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
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

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := h.service.Files.UploadFile(sessionID, uploadingFile, fileName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully uploaded", fileName)
	h.ts.ExecuteTemplate(w, "uploadGood.html", fileName)
}

// func newFileHandler(template *template.Template, service service.Service) *fileHandler {
// 	return &fileHandler{
// 		ts:      template,
// 		service: service,
// 	}
// }
