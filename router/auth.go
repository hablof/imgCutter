package router

func (r *Handler) GetNewSessionUUID() string {
	return r.service.Session.New().String()
}

func (r *Handler) AuthSession(id string) bool {
	_, ok := r.service.Session.Find(id)
	return ok
}
