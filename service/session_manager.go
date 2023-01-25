package service

import "github.com/google/uuid"

type Session struct {
	id uuid.UUID
}

func (s *Session) String() string {
	return s.id.String()
}

type sessionManager struct {
	sessions map[string]*Session
}

func (sm *sessionManager) Find(id string) (ses *Session, ok bool) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}
	ses, ok = sm.sessions[id]
	return ses, ok
}

func (sm *sessionManager) New() *Session {
	var ses Session
	for {
		ses = Session{
			id: uuid.New(),
		}
		if _, ok := sm.sessions[ses.id.String()]; !ok {
			break
		}
	}
	sm.sessions[ses.id.String()] = &ses
	return &ses
}

func (sm *sessionManager) GetAll() []string {
	out := make([]string, 0, len(sm.sessions))
	for _, v := range sm.sessions {
		out = append(out, v.String())
	}
	return out
}
