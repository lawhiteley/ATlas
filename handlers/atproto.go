package handlers

import (
	"ATlas/models"
	"context"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
)

func (s *Server) PutPinRecord(session models.Session) {
	// TODO: capture and handle err from GetSession
	sess, _ := s.Repository.GetSession(context.Background(), *session.DID, session.SessionID)
	client := &xrpc.Client{
		Host: "https://bsky.social", // TODO: configurable
		Auth: &xrpc.AuthInfo{
			AccessJwt:  sess.AccessToken,
			RefreshJwt: sess.RefreshToken,
			Did:        session.DID.String(),
		},
	}

	// TODO: log above output

	// TODO: handle Pin construct and failure scenario
	atproto.RepoPutRecord(context.Background(), client, &atproto.RepoPutRecord_Input{})
}
