package main

import (
	"ATlas/db"
	"ATlas/handlers"
	"context"
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
	app := &cli.Command{
		Name:   "ATlas",
		Action: initializeServer,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "bind",
				Aliases: []string{"b"},
				Usage:   "Server bind address",
				Value:   ":3000",
			},
			&cli.StringFlag{
				Name:    "hostname",
				Usage:   "Public hostname (will configure in localhost conditions if blank)",
				Sources: cli.EnvVars("CLIENT_HOSTNAME"),
			},
			&cli.StringFlag{
				Name:    "client-secret-key",
				Usage:   "Key for confidential client. P-256 in multibase encoding",
				Sources: cli.EnvVars("CLIENT_SECRET_KEY"),
			},
			&cli.StringFlag{
				Name:    "client-secret-key-id",
				Usage:   "Key ID for $CLIENT_SECRET_KEY",
				Value:   "primary",
				Sources: cli.EnvVars("CLIENT_SECRET_KEY_ID"),
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error("ATlas failed to start", "error", err)
		os.Exit(1)
	}
}

func initializeServer(cctx context.Context, cmd *cli.Command) error {
	bind := cmd.String("bind")
	mux := http.NewServeMux()

	// TODO: add confidential client support for deploy
	config := oauth.NewLocalhostConfig(
		fmt.Sprintf("http://127.0.0.1%s/auth/callback", bind),
		[]string{
			"atproto",
			"rpc:app.bsky.actor.getProfile?aud=did:web:api.bsky.app%23bsky_appview",
			"transition:generic",
		}, // TODO: extract
	)

	store, _ := db.NewSQLiteStore(&db.SQLiteConfig{
		Path:             "atlas_data.sqlite3",
		SessionExpiry:    time.Hour * 24 * 90,
		InactivityExpiry: time.Hour * 24 * 14,
		RequestExpiry:    time.Minute * 30,
	})

	oauthClient := oauth.NewClientApp(&config, store)

	srv := handlers.Server{
		Repository:  store,
		CookieStore: sessions.NewCookieStore([]byte("dfklsdjfkjldfjklsdfjlf")), // TODO: cctx.String("session-secret") from CLI arg
		Dir:         identity.DefaultDirectory(),
		OAuth:       oauthClient,
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	registerAppRoutes(mux, &srv)
	registerAuthRoutes(mux, &srv)

	if err := http.ListenAndServe(bind, mux); err != nil {
		slog.Error("failed to start listener", "err", err)
		return err
	}

	return nil
}

func registerAppRoutes(mux *http.ServeMux, s *handlers.Server) {
	mux.HandleFunc("GET /", s.Globe)

	mux.HandleFunc("POST /pin", s.NewPin)
	mux.HandleFunc("DELETE /pin", s.RemovePin)

}

func registerAuthRoutes(mux *http.ServeMux, s *handlers.Server) {

	// Outward-facing auth
	mux.HandleFunc("GET /auth/jwks.json", s.JWKS)
	mux.HandleFunc("GET /auth/client-metadata.json", s.ClientMetadata)
	mux.HandleFunc("GET /auth/callback", s.OAuthCallback)

	// User-facing auth
	mux.HandleFunc("GET /auth/login", s.OAuthLogin)
	mux.HandleFunc("POST /auth/login", s.OAuthLogin)
	mux.HandleFunc("GET /auth/logout", s.OAuthLogout)
}
