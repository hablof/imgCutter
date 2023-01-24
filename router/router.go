package router

import (
	"html/template"
	"imgCutter/service"
	"net/http"
	"path/filepath"
)

type FileHandler interface {
	CutFile(w http.ResponseWriter, r *http.Request)
	MainPage(w http.ResponseWriter, r *http.Request)
	DownloadFile(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
}

type Router struct {
	FileHandler
}

func NewRouter(s service.Service) (*Router, error) {
	ts, err := template.New("home.html").Funcs(template.FuncMap{
		"base": filepath.Base,
	}).ParseGlob("static/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Router{
		FileHandler: newFileHandler(ts, s),
	}, nil
}

func (h *Router) GetHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", h.DownloadFile)
	mux.HandleFunc("/", h.MainPage)
	mux.HandleFunc("/upload", h.UploadFile)
	mux.HandleFunc("/cut", h.CutFile)
	return mux
}
