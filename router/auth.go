package router

import (
	"log"
	"net/http"
)

func (h *Handler) TerminateSession(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value(ctxSessionKey).(string)
	if !ok {
		log.Printf("unable to get context value")
		w.WriteHeader(http.StatusInternalServerError)
	}

	session, ok := h.service.Session.Find(sessionID)
	if !ok {
		log.Printf("Session not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.service.Session.TerminateSession(session); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = h.templates.ExecuteTemplate(w, "terminateGood.html", nil)
}
