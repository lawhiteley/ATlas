package handlers

import (
	"ATlas/db"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/gorilla/sessions"
)

type Server struct {
	Repository  *db.SQLiteStore
	CookieStore *sessions.CookieStore
	Dir         identity.Directory
	OAuth       *oauth.ClientApp
}

func (s *Server) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != "POST" {
		bskyAuth, _ := s.OAuth.StartAuthFlow(ctx, "https://bsky.social")
		http.Redirect(w, r, bskyAuth, http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Errorf("Failed to parse form: %w", err).Error(), http.StatusBadRequest)
		return
	}

	username, _ := strings.CutPrefix(r.PostFormValue("username"), "@")
	slog.Info("Auth flow started for client", "clientID", s.OAuth.Config.ClientID, "callbackURL", s.OAuth.Config)

	redirectURL, err := s.OAuth.StartAuthFlow(ctx, username)

	if err != nil {
		var oauthErr = fmt.Errorf("OAuth failure: %w", err).Error()
		slog.Error(oauthErr)
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, oauthErr)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (s *Server) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := s.OAuth.ProcessCallback(ctx, r.URL.Query())
	if err != nil {
		var callbackErr = fmt.Errorf("Failed to process OAuth callback: %w", err).Error()
		slog.Error(callbackErr)
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, callbackErr)

		return
	}

	session, err := s.OAuth.ResumeSession(ctx, data.AccountDID, data.SessionID)
	if err != nil {
		slog.Error("Unauthenticated", "DID", data.AccountDID)
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, "Unauthenticated")

		return
	}

	c := session.APIClient()
	var getSession struct {
		Handle string `json:"handle"`
	}
	if err := c.Get(ctx, "com.atproto.server.getSession", nil, &getSession); err != nil {
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, err.Error())
		return
	}

	userDID := data.AccountDID.String()

	var getProfile struct {
		Avatar      string `json:"avatar"`
		DisplayName string `json:"displayName"`
	}
	profile := map[string]any{"actor": userDID}
	if err := c.Get(ctx, "app.bsky.actor.getProfile", profile, &getProfile); err != nil {
		slog.Error("Failed to get profile for user", "DID", userDID)
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, err.Error())
		return
	}

	cookie, _ := s.CookieStore.Get(r, "oauth-session")
	cookie.Values["account_did"] = userDID
	cookie.Values["display_name"] = getProfile.DisplayName

	if getProfile.Avatar != "" {
		cookie.Values["avatar"] = getProfile.Avatar
	} else {
		cookie.Values["avatar"] = "/static/img/default_avatar.png"
	}

	cookie.Values["session_id"] = data.SessionID
	cookie.Values["handle"] = getSession.Handle
	if err := cookie.Save(r, w); err != nil {
		slog.Error("Failed to save cookie for user", "DID", userDID)
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, err.Error())
		return
	}

	slog.Info("Successful login", "did", userDID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) OAuthLogout(w http.ResponseWriter, r *http.Request) {
	session := s.getDID(r)

	if session.DID != nil {
		if err := s.OAuth.Logout(r.Context(), *session.DID, session.SessionID); err != nil {
			slog.Error("Logout failed", "did", session.DID, "err", err)
		}
	}

	sess, _ := s.CookieStore.Get(r, "oauth-session")
	sess.Values = make(map[any]any)
	err := sess.Save(r, w)
	if err != nil {
		slog.Error("Failed to delete session", "did", session.DID, "err", err)
		renderDefaultGlobe(r.Context(), w, s.getDID(r), nil, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "oauth-session", Value: "", Path: "/", MaxAge: -1})

	slog.Info("Successful logout", "did", session.DID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) ClientMetadata(w http.ResponseWriter, r *http.Request) {
	slog.Info("Client Metadata request", "url", r.URL, "host", r.Host)

	meta := s.OAuth.Config.ClientMetadata()
	if s.OAuth.Config.IsConfidential() {
		meta.JWKSURI = strPtr(fmt.Sprintf("https://%s/oauth/jwks.json", r.Host))
	}
	meta.ClientName = strPtr("ATlas")
	meta.ClientURI = strPtr(fmt.Sprintf("https://%s", r.Host))

	if err := meta.Validate(s.OAuth.Config.ClientID); err != nil {
		slog.Error("Failed when validating client metadata", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) JWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body := s.OAuth.Config.PublicJWKS()
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func strPtr(raw string) *string {
	return &raw
}
