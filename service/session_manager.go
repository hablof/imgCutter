package service

import "github.com/google/uuid"

func (fm *fileManager) Find(id string) (ses *Session, ok bool) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}
	ses, ok = fm.sessions[id]
	return ses, ok
}

func (sm *fileManager) New() *Session {
	var ses Session
	for {
		ses = Session{
			id:    uuid.New(),
			files: map[string]myFile{},
		}
		if _, ok := sm.sessions[ses.id.String()]; !ok {
			break
		}
	}
	sm.sessions[ses.id.String()] = &ses
	return &ses
}

func (sm *fileManager) GetAll() []string {
	out := make([]string, 0, len(sm.sessions))
	for _, v := range sm.sessions {
		out = append(out, v.String())
	}
	return out
}
