package router

import (
	"html/template"
	"imgCutter/service"
	"net/http"
	"path/filepath"
)

type Handler struct {
	ts      *template.Template
	service service.Service
}

func NewRouter(s service.Service) (*Handler, error) {
	ts, err := template.New("home.html").Funcs(template.FuncMap{
		"base": filepath.Base,
	}).ParseGlob("static/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{
		ts:      ts,
		service: s,
	}, nil
}

func (h *Handler) GetHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", h.DownloadFile)
	mux.HandleFunc("/", h.MainPage)
	mux.HandleFunc("/upload", h.UploadFile)
	mux.HandleFunc("/cut", h.CutFile)
	mux.HandleFunc("/favicon.ico", h.favicon)
	handler := h.Logging(h.ManageSession(mux.ServeHTTP))
	return handler
}
