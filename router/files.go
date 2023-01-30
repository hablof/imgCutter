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
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	fileName := r.PostForm.Get("fileName")

	dX, err := strconv.Atoi(r.PostForm.Get("dX"))
	if err != nil {
		log.Printf("error parsing dX: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	dY, err := strconv.Atoi(r.PostForm.Get("dY"))
	if err != nil {
		log.Printf("error parsing dY: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	log.Printf("cutting file: %v, dX: %v px, dY: %v px", filepath.Base(fileName), dX, dY)

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	session, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("Session not found")
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if err := h.service.Files.CutFile(session, fileName, dX, dY); err != nil {
		log.Printf("error processing img: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	log.Printf("file %s succsesfully cut", filepath.Base(fileName))

	b := bytes.Buffer{}

	if err := h.templates.ExecuteTemplate(&b, "cutGood.html", fileName); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, b.String())
}

func (h *Handler) MainPage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	s, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("Session not found")
		w.WriteHeader(http.StatusNotFound)

		return
	}

	filesList, err := h.service.Files.GetFiles(s)
	if err != nil {
		log.Printf("unable to get files list")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	b := bytes.Buffer{}

	if err := h.templates.ExecuteTemplate(&b, "home.html", filesList); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, b.String())
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
		return
	}

	s, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("Session not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	archiveName, err := h.service.Files.GetArchiveName(s, fileName)
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
		return
	}

	s, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("Session not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.service.Files.UploadFile(s, uploadingFile, fileName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully uploaded", fileName)
	_ = h.templates.ExecuteTemplate(w, "uploadGood.html", fileName)
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("err parsing form: %v", err)
		return
	}

	fileName := r.PostForm.Get("fileName")

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("Session not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.service.Files.DeleteFile(session, fileName); err != nil {
		log.Printf("unable to delete files")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("file %s succsesfully deleted", fileName)
	_ = h.templates.ExecuteTemplate(w, "deleteGood.html", fileName)
}

// func newFileHandler(template *template.Template, service service.Service) *fileHandler {
// 	return &fileHandler{
// 		ts:      template,
// 		service: service,
// 	}
// }
