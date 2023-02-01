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
		fmt.Fprint(w, "Bad Request")

		return
	}

	if !r.PostForm.Has("fileName") {
		log.Printf(`request form missing field "fileName"`)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}
	fileName := r.PostForm.Get("fileName")

	dX, err := strconv.Atoi(r.PostForm.Get("dX"))
	if err != nil {
		log.Printf("error parsing dX: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}

	dY, err := strconv.Atoi(r.PostForm.Get("dY"))
	if err != nil {
		log.Printf("error parsing dY: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}

	log.Printf("cutting file: %v, dX: %v px, dY: %v px", filepath.Base(fileName), dX, dY)

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	session, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("session not found")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Bad Session")

		return
	}

	if err := h.service.Files.CutFile(session, fileName, dX, dY); err != nil {
		log.Printf("error processing img: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

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
	w.Write(b.Bytes())
}

func (h *Handler) MainPage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	s, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("session not found")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Bad Session")

		return
	}

	filesList, err := h.service.Files.GetFiles(s)
	if err != nil {
		log.Printf("unable to get files list")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

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
	w.Write(b.Bytes())
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("err parsing form: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if !r.PostForm.Has("fileName") {
		log.Printf(`request form missing field "fileName"`)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}
	fileName := r.PostForm.Get("fileName")
	log.Printf("downloading archive of: %v", fileName)

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	s, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("session not found")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Bad Session")

		return
	}

	archiveName, err := h.service.Files.GetArchiveName(s, fileName)
	if err != nil {
		log.Printf("error getting archive name: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

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

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	s, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("session not found")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Bad Session")

		return
	}

	uploadingFile, fileHeader, err := r.FormFile("uploadingFile")
	if err != nil {
		log.Printf("FormFile error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}
	defer uploadingFile.Close()

	contentType := fileHeader.Header.Get("content-type")
	fileName := fileHeader.Filename

	if !(contentType == "image/jpeg" || contentType == "image/png") {
		log.Printf("invalid fileHeader content-type: %s", contentType)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "file must be .jpg (or .png)")

		return
	}

	if err := h.service.Files.UploadFile(s, uploadingFile, fileName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	b := bytes.Buffer{}

	if err := h.templates.ExecuteTemplate(&b, "uploadGood.html", fileName); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	log.Printf("file %s succsesfully uploaded", fileName)
	w.WriteHeader(http.StatusOK)
	w.Write(b.Bytes())
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	session, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("session not found")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Bad Session")

		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("err parsing form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}

	if !r.PostForm.Has("fileName") {
		log.Printf(`request form missing field "fileName"`)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")

		return
	}
	fileName := r.PostForm.Get("fileName")

	if err := h.service.Files.DeleteFile(session, fileName); err != nil {
		log.Printf("unable to delete files")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	b := bytes.Buffer{}

	if err := h.templates.ExecuteTemplate(&b, "deleteGood.html", fileName); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")

		return
	}

	log.Printf("file %s succsesfully deleted", fileName)
	w.WriteHeader(http.StatusOK)
	w.Write(b.Bytes())
}
