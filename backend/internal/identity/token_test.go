package identity

import (
	"bytes"
	"testing"
)

func TestCryptoSessionTokenGeneratorRoundTrip(t *testing.T) {
	tokens := NewCryptoSessionTokenGenerator()
	rawToken, generatedHash, err := tokens.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if rawToken == "" || bytes.Contains(generatedHash, []byte(rawToken)) {
		t.Fatalf("unexpected raw token or hash")
	}
	recomputedHash, err := tokens.Hash(rawToken)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !bytes.Equal(generatedHash, recomputedHash) || len(generatedHash) != 32 {
		t.Fatalf("token digest mismatch")
	}
	if _, err := tokens.Hash("invalid"); err == nil {
		t.Fatal("Hash(invalid) error = nil")
	}
}

func TestCryptoSessionTokensAreUnique(t *testing.T) {
	tokens := NewCryptoSessionTokenGenerator()
	first, _, err := tokens.New()
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := tokens.New()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two generated session tokens were equal")
	}
}
