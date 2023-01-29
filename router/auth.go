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

	if err := h.service.Session.TerminateSession(sessionID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = h.templates.ExecuteTemplate(w, "terminateGood.html", nil)
}
