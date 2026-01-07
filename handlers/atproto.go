package handlers

import (
	"ATlas/models"
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/lex/util"
)

func (s *Server) PutPinRecord(session models.Session, pin models.Pin) string {
	client, _ := s.OAuth.ResumeSession(context.Background(), *session.DID, session.SessionID)
	at := client.APIClient()

	pinRecord := &atproto.RepoPutRecord_Input{
		Collection: "io.whiteley.ATlas.pin",
		Repo:       session.DID.String(),
		Rkey:       "self",
		Record: &util.LexiconTypeDecoder{Val: &models.AtlasPinRecord{
			Did:         session.DID.String(),
			PlacedAt:    time.Now().UTC().Format(time.RFC3339),
			Longitude:   strconv.FormatFloat(pin.Latitude, 'f', -1, 64),
			Latitude:    strconv.FormatFloat(pin.Longitude, 'f', -1, 64),
			Description: pin.Description,
			Website:     &pin.Website,
		}},
	}

	slog.Debug("Atproto createRecord payload", "record", pinRecord)

	var response struct {
		Uri syntax.ATURI `json:"uri"`
	}

	if err := at.Post(context.Background(), "com.atproto.repo.createRecord", pinRecord, &response); err != nil {
		slog.Error("PutRecord request failed", "err", err)

	}

	pinURI := response.Uri.String()
	slog.Info("Atproto createRecord request successful", "pinURI", pinURI)

	return pinURI
}

func (s *Server) DeletePinRecord(session models.Session) {
	client, _ := s.OAuth.ResumeSession(context.Background(), *session.DID, session.SessionID)
	at := client.APIClient()

	record := map[string]any{
		"collection": "io.whiteley.ATlas.pin",
		"repo":       session.DID.String(),
		"rkey":       "self",
	}

	slog.Debug("Atproto deleteRecord payload", "record", record)

	if err := at.Post(context.Background(), "com.atproto.repo.deleteRecord", record, nil); err != nil {
		slog.Error("Atproto createRecord request failed", "err", err)
	}

	slog.Info("Atproto createRecord request successful", "DID", session.DID.String())
}
