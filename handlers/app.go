package handlers

import (
	"ATlas/components"
	"ATlas/models"
	"context"
	"log/slog"
	"net/http"
)

func (s *Server) Globe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pins, perr := s.Repository.GetPins(r.Context())

	if perr != nil {
		slog.Info("Initial pin load failed", "err", perr)
	}

	session := s.getDID(r)
	didMap := ToDidMap(pins)

	if session.DID == nil {
		renderDefaultGlobe(ctx, w, session, didMap, "")
		return
	}

	_, err := s.OAuth.ResumeSession(ctx, *session.DID, session.SessionID)
	if err != nil {
		slog.Warn("Session resume failed for user", "DID", session.DID, "sessionID", session.SessionID)
		renderDefaultGlobe(ctx, w, session, didMap, "")
		return
	}

	slog.Info("Globe loaded for user", "DID", session.DID, "handle", session.Handle, "name", session.Name)
	userDid := session.DID.String()
	userPin := didMap[userDid]
	v := components.Page(components.Atlas(userDid, didMap), components.Panel(true, session), &userPin, "")
	v.Render(ctx, w)
}

func renderDefaultGlobe(ctx context.Context, w http.ResponseWriter, session models.Session, didMap map[string]models.Pin, flash string) {
	v := components.Page(components.Atlas("", didMap), components.Panel(false, session), nil, flash)
	v.Render(ctx, w)
}

func ToDidMap(pins []models.Pin) map[string]models.Pin {
	didMap := make(map[string]models.Pin, len(pins))
	for _, pin := range pins {
		didMap[pin.Did] = pin
	}

	return didMap
}
