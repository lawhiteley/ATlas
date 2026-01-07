package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"log/slog"

	"github.com/MicahParks/jwkset"
)

const logFmt = "%s\nError: %s"

func main() {
	curve := elliptic.P256()
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		slog.Error("Failed to generate ECDSA/P-256 key", err)
		return
	}

	marshalOptions := jwkset.JWKMarshalOptions{
		Private: true,
	}
	metadata := jwkset.JWKMetadataOptions{
		KID: "ATlas-confidential-client",
		// TODO: set ALG? KEYOPS? USE?
	}
	options := jwkset.JWKOptions{
		Marshal:  marshalOptions,
		Metadata: metadata,
	}

	jwk, err := jwkset.NewJWKFromKey(priv, options)
	if err != nil {
		slog.Error("Failed to create JWK from key", err)
		return
	}

	json, err := json.MarshalIndent(jwk.Marshal(), "", "  ")
	if err != nil {
		slog.Error("Failed marshal JSON", err)
		return
	}

	println(string(json))
}
