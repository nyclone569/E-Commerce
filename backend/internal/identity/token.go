package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

const sessionTokenBytes = 32

type SessionTokenGenerator interface {
	New() (string, []byte, error)
	Hash(string) ([]byte, error)
}

type CryptoSessionTokenGenerator struct{}

func NewCryptoSessionTokenGenerator() *CryptoSessionTokenGenerator {
	return &CryptoSessionTokenGenerator{}
}

func (*CryptoSessionTokenGenerator) New() (string, []byte, error) {
	randomBytes := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}
	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	digest := sha256.Sum256([]byte(rawToken))
	return rawToken, digest[:], nil
}

func (*CryptoSessionTokenGenerator) Hash(rawToken string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil || len(decoded) != sessionTokenBytes {
		return nil, errors.New("invalid session token")
	}
	digest := sha256.Sum256([]byte(rawToken))
	return digest[:], nil
}
