package handlers

import (
	"ATlas/db"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/gorilla/sessions"
)

func (s *Server) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	os.Stdout.WriteString("DIRECT: Handler called\n")
	slog.Info("login hello")

	if r.Method != "POST" {
		bskyAuth, _ := s.OAuth.StartAuthFlow(ctx, "https://bsky.social")
		slog.Info("goin to bsky", "at", bskyAuth)
		http.Redirect(w, r, bskyAuth, http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Errorf("failed to parse form: %w", err).Error(), http.StatusBadRequest)
		return
	}

	username, _ := strings.CutPrefix(r.PostFormValue("username"), "@")
	slog.Info("login", "client_id", s.OAuth.Config.ClientID, "callback_url", s.OAuth.Config.CallbackURL)

	redirectURL, err := s.OAuth.StartAuthFlow(ctx, username)

	if err != nil {
		var oauthErr = fmt.Errorf("OAuth failure: %w", err).Error()
		slog.Error(oauthErr)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (s *Server) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := r.URL.Query()
	slog.Info("callback", "params", params)

	data, err := s.OAuth.ProcessCallback(ctx, r.URL.Query())
	if err != nil {
		var callbackErr = fmt.Errorf("failed to process callback: %w", err).Error()
		slog.Error(callbackErr)
		return
	}

	session, err := s.OAuth.ResumeSession(ctx, data.AccountDID, data.SessionID)
	if err != nil {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	c := session.APIClient()

	var getSession struct {
		Handle string `json:"handle"`
	}
	if err := c.Get(ctx, "com.atproto.server.getSession", nil, &getSession); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var getProfile struct {
		Avatar      string `json:"avatar"`
		DisplayName string `json:"displayName"`
	}
	profile := map[string]any{"actor": data.AccountDID.String()}
	if err := c.Get(ctx, "app.bsky.actor.getProfile", profile, &getProfile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("result", "profile", getProfile)

	cookie, _ := s.CookieStore.Get(r, "oauth-session")
	cookie.Values["account_did"] = data.AccountDID.String()
	cookie.Values["display_name"] = getProfile.DisplayName

	if getProfile.Avatar != "" {
		cookie.Values["avatar"] = getProfile.Avatar
	} else {
		cookie.Values["avatar"] = "/static/img/default_avatar.jpg"
	}

	cookie.Values["session_id"] = data.SessionID
	cookie.Values["handle"] = getSession.Handle
	if err := cookie.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("login success", "did", data.AccountDID.String())
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) OAuthLogout(w http.ResponseWriter, r *http.Request) {
	session := s.getDID(r)
	if session.DID != nil {
		if err := s.OAuth.Logout(r.Context(), *session.DID, session.SessionID); err != nil {
			slog.Error("logout failed", "did", session.DID, "err", err)
		}
	}

	sess, _ := s.CookieStore.Get(r, "oauth-session")
	sess.Values = make(map[any]any)
	err := sess.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "oauth-session", Value: "", Path: "/", MaxAge: -1})

	slog.Info("logged out", "did", session.DID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) ClientMetadata(w http.ResponseWriter, r *http.Request) {
	slog.Info("client metadata request", "url", r.URL, "host", r.Host)

	meta := s.OAuth.Config.ClientMetadata()
	if s.OAuth.Config.IsConfidential() {
		meta.JWKSURI = strPtr(fmt.Sprintf("https://%s/oauth/jwks.json", r.Host))
	}
	meta.ClientName = strPtr("ATlas")
	meta.ClientURI = strPtr(fmt.Sprintf("https://%s", r.Host))

	// internal consistency check
	if err := meta.Validate(s.OAuth.Config.ClientID); err != nil {
		slog.Error("validating client metadata", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TODO: move?
func strPtr(raw string) *string {
	return &raw
}

func (s *Server) JWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body := s.OAuth.Config.PublicJWKS()
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Server struct {
	Repository  *db.SQLiteStore
	CookieStore *sessions.CookieStore
	Dir         identity.Directory
	OAuth       *oauth.ClientApp
}
