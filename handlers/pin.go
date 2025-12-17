package handlers

import (
	"encoding/json"
	"net/http"
)

type Pin struct {
	Did         string
	Longitude   float64
	Latitude    float64
	Handle      string
	Description string
}

func (s *Server) NewPin(w http.ResponseWriter, r *http.Request) {
	var pin Pin
	if err := json.NewDecoder(r.Body).Decode(&pin); err != nil {
		http.Error(w, "Bad pin", http.StatusBadRequest)
		return
	}
	// for now just add every pin
}
