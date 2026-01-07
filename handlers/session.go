package handlers

import (
	"ATlas/models"
	"net/http"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

func (s *Server) getDID(r *http.Request) models.Session {
	current, _ := s.CookieStore.Get(r, "oauth-session")

	var session models.Session
	accountDID, ok := current.Values["account_did"].(string)
	if !ok || accountDID == "" {
		return session
	}

	did, err := syntax.ParseDID(accountDID)
	if err != nil {
		return session
	}

	sessionID, _ := current.Values["session_id"].(string)
	handle, _ := current.Values["handle"].(string)
	avatar, _ := current.Values["avatar"].(string)
	name, _ := current.Values["display_name"].(string)

	return models.Session{
		DID:       &did,
		SessionID: sessionID,
		Handle:    handle,
		Avatar:    avatar,
		Name:      name,
	}
}

func (s *Server) setFlash(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := s.CookieStore.Get(r, "oauth-session")

	session.AddFlash(message)
	session.Save(r, w)
}

func (s *Server) getFlash(w http.ResponseWriter, r *http.Request) (string, bool) {
	session, _ := s.CookieStore.Get(r, "oauth-session")

	flashes := session.Flashes()
	session.Save(r, w)

	if len(flashes) > 0 {
		return flashes[0].(string), true
	}
	return "", false
}
