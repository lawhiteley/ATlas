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

func (s *Server) PutPinRecord(session models.Session, pin models.Pin) {
	// TODO: refactor err handling etc..

	client, _ := s.OAuth.ResumeSession(context.Background(), *session.DID, session.SessionID)
	at := client.APIClient()

	pinRecord := &atproto.RepoPutRecord_Input{
		Collection: "io.whiteley.luke.ATlas.pin",
		Repo:       session.DID.String(),
		Rkey:       "self",
		Record: &util.LexiconTypeDecoder{Val: &models.ATlasPin{
			Did:         session.DID.String(),
			PlacedAt:    time.Now().UTC().Format(time.RFC3339),
			Longitude:   strconv.FormatFloat(pin.Latitude, 'f', -1, 64),
			Latitude:    strconv.FormatFloat(pin.Longitude, 'f', -1, 64),
			Description: pin.Description,
			Website:     &pin.Website,
		}},
	}

	slog.Info("atproto prepped", "record", pinRecord)
	var response struct { // TODO: persist and surface this
		Uri syntax.ATURI `json:"uri"`
	}

	if err := at.Post(context.Background(), "com.atproto.repo.createRecord", pinRecord, &response); err != nil {
		slog.Info("atproto bad", "err", err)

	}
	slog.Info("atproto good", "resp", response)

}

func (s *Server) DeletePinRecord(session models.Session) {
	// TODO: refactor err handling etc..
	// TODO: check DID equality?

	client, _ := s.OAuth.ResumeSession(context.Background(), *session.DID, session.SessionID)
	at := client.APIClient()

	record := map[string]any{
		"collection": "io.whiteley.luke.ATlas.pin",
		"repo":       session.DID.String(),
		"rkey":       "self",
	}

	slog.Info("atproto prepped", "record", record)

	if err := at.Post(context.Background(), "com.atproto.repo.deleteRecord", record, nil); err != nil {
		slog.Info("atproto bad", "err", err)

	}

	slog.Info("atproto good")

}
