package router

import (
	"html/template"
	"net/http"
	"path/filepath"

	"imgcutter/service"
)

type Handler struct {
	templates *template.Template
	service   service.Service
}

func NewRouter(s service.Service) (*Handler, error) {
	templates, err := template.New("home.html").Funcs(template.FuncMap{
		"base": filepath.Base,
	}).ParseGlob("static/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{
		templates: templates,
		service:   s,
	}, nil
}

func (h *Handler) GetHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.MainPage)
	mux.HandleFunc("/cut", h.CutFile)
	mux.HandleFunc("/download", h.DownloadFile)
	mux.HandleFunc("/delete", h.DeleteFile)
	mux.HandleFunc("/favicon.ico", h.favicon)
	mux.HandleFunc("/terminate", h.TerminateSession)
	mux.HandleFunc("/upload", h.UploadFile)
	handler := h.Logging(h.ManageSession(mux.ServeHTTP))
	return handler
}
