package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type ctxStr string

const (
	sessionID            = "SESSID"
	ctxSessionKey ctxStr = "sessionID"
	cookieLife           = 300 // in seconds
)

func (h *Handler) Logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println()
		log.Printf("request: %s %s %s", r.Host, r.Method, r.URL.String())
		f(w, r)
		//log.Print(h.service.Session.GetAll())
	}
}

func (h *Handler) ManageSession(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := h.checkSessionCookie(r)
		if !ok {
			log.Printf("creating new session")
			session = h.service.Session.New().String() // side-effect: new session entry in service.Session
			h.service.Files.NewSessionFiles(session)   // vsrato
			h.setSessionCookie(session, w)
		}

		log.Printf("working session: %s", session)

		//помещаем сессию в контекст запроса
		ctxr := r.WithContext(context.WithValue(context.Background(), ctxSessionKey, session))
		f(w, ctxr)
	}
}

func (h *Handler) checkSessionCookie(r *http.Request) (sessionUUID string, ok bool) {
	sessionCookie, err := r.Cookie(sessionID)
	//если куки инвалидны -- создаём новую сессию и пишем куки в ответ
	if err != nil {
		log.Printf("session cookie not found")
		return "", false
	}
	// если куки найдены, берём значение
	session := sessionCookie.Value

	// проверяем наличие сессии
	_, ok = h.service.Session.Find(session)
	if !ok {
		log.Printf("session not found")
		return "", false
	}
	return session, true
}

func (*Handler) setSessionCookie(session string, w http.ResponseWriter) {
	newCookie := http.Cookie{
		Name:  sessionID,
		Value: session,
		Path:  "/",

		MaxAge:   cookieLife, // in seconds
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, &newCookie)
	log.Printf("sent cookie: %s", newCookie.String())
}
