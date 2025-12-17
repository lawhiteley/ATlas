package models

import "github.com/bluesky-social/indigo/atproto/syntax"

type Session struct {
	DID       *syntax.DID
	SessionID string
	Handle    string
	Avatar    string
	Name      string
}
