package main

import (
	"ATlas/components"
	"ATlas/handlers"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// apply middleware for session, logging etc...
	// move? rename?
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index := components.Page("Luke", components.Atlas(), components.Panel(false, "law.png", "Luke"))

		index.Render(r.Context(), w)
	})

	registerAuthRoutes(mux)

	logger.Info("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}

func registerAuthRoutes(mux *http.ServeMux) {

	// Outward-facing auth
	mux.HandleFunc("GET /auth/jwks.json", handlers.JWKS)
	mux.HandleFunc("GET /auth/client-metadata.json", handlers.ClientMetadata)
	mux.HandleFunc("GET /auth/callback", handlers.OAuthCallback)

	// User-facing auth
	mux.HandleFunc("GET /auth/login", handlers.OAuthLogin)
	mux.HandleFunc("POST /auth/login", handlers.OAuthLogin)
	mux.HandleFunc("GET /auth/logout", handlers.OAuthLogout)
}
