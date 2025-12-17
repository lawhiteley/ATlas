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
