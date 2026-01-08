package main

import (
	"ATlas/db"
	"ATlas/handlers"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/bluesky-social/indigo/atproto/atcrypto"
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
				Value:   ":8080",
			},
			&cli.StringFlag{
				Name:    "hostname",
				Usage:   "Public hostname (will configure in localhost conditions if blank)",
				Sources: cli.EnvVars("CLIENT_HOSTNAME"),
			},
			&cli.StringFlag{
				Name:    "client-secret-key-id",
				Usage:   "Key ID for $CLIENT_SECRET_KEY",
				Value:   "primary",
				Sources: cli.EnvVars("CLIENT_SECRET_KEY_ID"),
			},
			&cli.StringFlag{
				Name:    "client-secret-key",
				Usage:   "Key for confidential client. P-256 in multibase encoding",
				Sources: cli.EnvVars("CLIENT_SECRET_KEY"),
			},
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Log verbosity",
				Value:   "debug",
				Sources: cli.EnvVars("LOG_LEVEL"),
			},
			&cli.StringFlag{
				Name:    "session-secret",
				Usage:   "Key for cookie security",
				Value:   "change-this-please",
				Sources: cli.EnvVars("SESSION_SECRET"),
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

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: getLogLevel(cmd)})
	slog.SetDefault(slog.New(handler))

	var config oauth.ClientConfig
	hostname := cmd.String("hostname")
	scope := []string{
		"atproto",
		"rpc:app.bsky.actor.getProfile?aud=did:web:api.bsky.app%23bsky_appview",
		"repo:io.whiteley.ATlas.pin",
	}

	if hostname == "" {
		config = oauth.NewLocalhostConfig(fmt.Sprintf("http://127.0.0.1%s/auth/callback", bind), scope)
	} else {
		config = oauth.NewPublicConfig(
			fmt.Sprintf("https://%s/auth/client-metadata.json", hostname),
			fmt.Sprintf("https://%s/auth/callback", hostname),
			scope,
		)
	}

	if hostname != "" && cmd.String("client-secret-key") != "" {
		priv, err := atcrypto.ParsePrivateMultibase(cmd.String("client-secret-key"))
		if err != nil {
			return err
		}
		if err := config.SetClientSecret(priv, cmd.String("client-secret-key-id")); err != nil {
			return err
		}
	}

	store, _ := db.NewSQLiteStore(&db.SQLiteConfig{
		Path:             "atlas_data.sqlite3",
		SessionExpiry:    time.Hour * 24 * 5,
		InactivityExpiry: time.Hour * 24 * 3,
		RequestExpiry:    time.Minute * 30,
	})

	oauthClient := oauth.NewClientApp(&config, store)

	srv := handlers.Server{
		Repository:  store,
		CookieStore: sessions.NewCookieStore([]byte(cmd.String("session-secret"))),
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
	// Atproto-facing auth
	mux.HandleFunc("GET /auth/jwks.json", s.JWKS)
	mux.HandleFunc("GET /auth/client-metadata.json", s.ClientMetadata)
	mux.HandleFunc("GET /auth/callback", s.OAuthCallback)

	// User-facing auth
	mux.HandleFunc("GET /auth/login", s.OAuthLogin)
	mux.HandleFunc("POST /auth/login", s.OAuthLogin)
	mux.HandleFunc("GET /auth/logout", s.OAuthLogout)
}

func getLogLevel(cmd *cli.Command) slog.Level {
	logLevels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	levelFromEnv := cmd.String("LOG_LEVEL")

	logLevel, ok := logLevels[strings.ToLower(levelFromEnv)]
	if !ok {
		logLevel = slog.LevelInfo
	}

	return logLevel
}
