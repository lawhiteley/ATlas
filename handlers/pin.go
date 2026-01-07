package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"ATlas/models"

	"github.com/go-playground/validator/v10"
)

func (s *Server) NewPin(w http.ResponseWriter, r *http.Request) {
	validator := validator.New()
	response := struct {
		PinData *models.Pin `json:"pinData"`
		Error   string      `json:"error,omitempty"`
	}{}

	if err := r.ParseForm(); err != nil {
		slog.Warn("Form validation failed", "err", err)
		response.Error = err.Error()
		return
	} else {
		longitude, _ := strconv.ParseFloat(r.FormValue("longitude"), 64)
		latitude, _ := strconv.ParseFloat(r.FormValue("latitude"), 64)
		description := r.FormValue("description")
		website := r.FormValue("website")

		session := s.getDID(r)

		pin := &models.Pin{
			Did:         string(*s.getDID(r).DID),
			Longitude:   longitude,
			Latitude:    latitude,
			Description: description,
			Website:     website,
			Name:        session.Name,
			Handle:      session.Handle,
			Avatar:      session.Avatar,
		}

		err := validator.Struct(pin)
		if err != nil {
			slog.Warn("Validation failed for pin", "err", err)
			response.Error = fmt.Sprintf("Pin validation failed: %s", err)
		} else {
			uri := s.PutPinRecord(session, *pin)
			pin.Uri = uri

			s.Repository.SavePin(r.Context(), *pin)
			response.PinData = pin
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (s *Server) RemovePin(w http.ResponseWriter, r *http.Request) {
	session := s.getDID(r)

	s.DeletePinRecord(session)
	err := s.Repository.DeletePin(r.Context(), session.DID.String())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
