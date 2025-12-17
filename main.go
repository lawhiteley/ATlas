package main

import (
	"ATlas/components"
	"ATlas/db"
	"ATlas/handlers"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/gorilla/sessions"
	"github.com/urfave/cli/v3"
)

func main() {
	app := cli.App{
		Name:   "oauth-web-demo",
		Usage:  "atproto OAuth web server demo",
		Action: initializeServer,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "session-secret",
				Usage:    "random string/token used for session cookie security",
				Required: true,
				EnvVars:  []string{"SESSION_SECRET"},
			},
			&cli.StringFlag{
				Name:    "hostname",
				Usage:   "public host name for this client (if not localhost dev mode)",
				EnvVars: []string{"CLIENT_HOSTNAME"},
			},
			&cli.StringFlag{
				Name:    "client-secret-key",
				Usage:   "confidential client secret key. should be P-256 private key in multibase encoding",
				EnvVars: []string{"CLIENT_SECRET_KEY"},
			},
			&cli.StringFlag{
				Name:    "client-secret-key-id",
				Usage:   "key id for client-secret-key",
				Value:   "primary",
				EnvVars: []string{"CLIENT_SECRET_KEY_ID"},
			},
		},
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

func initializeServer(cctx *cli.Context) error {
	bind := ":3000" // TODO: check for in-use port and panic
	mux := http.NewServeMux()

	// TODO: add confidential client support for deploy
	config := oauth.NewLocalhostConfig(
		fmt.Sprintf("http://127.0.0.1%s/auth/callback", bind),
		[]string{"atproto", "repo:app.bsky.feed.post?action=create"}, // TODO: switch out for pin lexicon later
	)

	store, _ := db.NewSQLiteStore(&db.SQLiteConfig{
		Path:             "oauth_sessions.sqlite3",
		SessionExpiry:    time.Hour * 24 * 90,
		InactivityExpiry: time.Hour * 24 * 14,
		RequestExpiry:    time.Minute * 30,
	})

	oauthClient := oauth.NewClientApp(&config, store)
	srv := handlers.Server{
		CookieStore: sessions.NewCookieStore([]byte("dfklsdjfkjldfjklsdfjlf")), // TODO: cctx.String("session-secret") from CLI arg
		Dir:         identity.DefaultDirectory(),
		OAuth:       oauthClient,
	}

	registerAppRoute(mux)
	registerAuthRoutes(mux, &srv)

	if err := http.ListenAndServe(bind, mux); err != nil {
		slog.Error("failed to start listener", "err", err)
		return err
	}

	return nil
}

func registerAppRoute(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index := components.Page("Luke", components.Atlas(), components.Panel(false, "law.png", "Luke"))

		index.Render(r.Context(), w)
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func registerAuthRoutes(mux *http.ServeMux, s *handlers.Server) {

	// Outward-facing auth
	mux.HandleFunc("GET /oauth/jwks.json", s.JWKS)
	mux.HandleFunc("GET /oauth/client-metadata.json", s.ClientMetadata)
	mux.HandleFunc("GET /oauth/callback", s.OAuthCallback)

	// User-facing auth
	mux.HandleFunc("GET /auth/login", s.OAuthLogin)
	mux.HandleFunc("POST /auth/login", s.OAuthLogin)
	mux.HandleFunc("GET /auth/logout", s.OAuthLogout)
}
