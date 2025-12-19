package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"ATlas/models"
)

// add GET /pin handler for streaming pins as map scrolls

func (s *Server) NewPin(w http.ResponseWriter, r *http.Request) {

	// for now just add every pin
	// eventually
	// reject if DID has a pin younger than 1 month
	if err := r.ParseForm(); err != nil {
		slog.Info("err", "err", err)

		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: validate these
	longitude, _ := strconv.ParseFloat(r.FormValue("longitude"), 64)
	latitude, _ := strconv.ParseFloat(r.FormValue("latitude"), 64)
	description := r.FormValue("description")

	slog.Info("placed at", "lat", latitude, "long", longitude)
	session := s.getDID(r)

	pin := &models.Pin{
		Did:         string(*s.getDID(r).DID),
		Longitude:   longitude,
		Latitude:    latitude,
		Description: description,
		Name:        session.Name,
		Handle:      session.Handle,
		Avatar:      session.Avatar,
	}

	s.Repository.SavePin(r.Context(), *pin)
	// TODO: persist

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pin)
}
