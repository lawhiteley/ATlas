package handlers

import (
	"ATlas/components"
	"ATlas/models"
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

	didMap := ToDidMap(pins)
	if session.DID == nil {
		// TODO: default values
		v := components.Page(components.Atlas("", didMap), components.Panel(false, session), nil)
		v.Render(ctx, w)
		return
	}

	_, err := s.OAuth.ResumeSession(ctx, *session.DID, session.SessionID)
	if err != nil {
		// oauth failed
	}

	userDid := session.DID.String()
	userPin := didMap[userDid]
	v := components.Page(components.Atlas(userDid, didMap), components.Panel(true, session), &userPin)
	v.Render(ctx, w)
}

func ToDidMap(pins []models.Pin) map[string]models.Pin {
	didMap := make(map[string]models.Pin, len(pins))
	for _, pin := range pins {
		didMap[pin.Did] = pin
	}

	return didMap
}
