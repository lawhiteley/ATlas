package handlers

import (
	"ATlas/components"
	"log/slog"
	"net/http"
)

func (s *Server) Globe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pins, perr := s.Repository.GetPins(r.Context())

	if perr != nil {
		slog.Info("failed to load stored pins")
	}

	slog.Info("pins", "p", pins)

	session := s.getDID(r)
	slog.Info("result", "did", session.DID, "sesh", session.Avatar, "handle", session.Handle, "name", session.Name)
	if session.DID == nil {
		v := components.Page("", components.Atlas(pins), components.Panel(false, session))
		v.Render(ctx, w)
		return
	}

	_, err := s.OAuth.ResumeSession(ctx, *session.DID, session.SessionID)
	if err != nil {
		// oauth failed
	}
	v := components.Page("", components.Atlas(pins), components.Panel(true, session))
	v.Render(ctx, w)

}
