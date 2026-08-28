package identity

import (
	"strings"
	"testing"
)

func TestArgon2idHasherHashAndVerify(t *testing.T) {
	hasher := NewArgon2idHasher(testArgon2Parameters())
	password := "correct horse battery staple"
	encoded, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if strings.Contains(encoded, password) || !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Fatalf("Hash() returned unexpected encoding: %q", encoded)
	}

	valid, err := hasher.Verify(encoded, password)
	if err != nil || !valid {
		t.Fatalf("Verify(correct) = %v, %v", valid, err)
	}
	valid, err = hasher.Verify(encoded, "wrong password that is long enough")
	if err != nil || valid {
		t.Fatalf("Verify(wrong) = %v, %v", valid, err)
	}
}

func TestArgon2idHasherUsesNFCNormalization(t *testing.T) {
	hasher := NewArgon2idHasher(testArgon2Parameters())
	encoded, err := hasher.Hash("Cafe\u0301 has a sufficiently long phrase")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	valid, err := hasher.Verify(encoded, "Café has a sufficiently long phrase")
	if err != nil || !valid {
		t.Fatalf("Verify(NFC equivalent) = %v, %v", valid, err)
	}
}

func TestArgon2idHasherRejectsMalformedRecord(t *testing.T) {
	hasher := NewArgon2idHasher(testArgon2Parameters())
	if _, err := hasher.Verify("not-a-password-hash", "anything"); err == nil {
		t.Fatal("Verify() error = nil, want malformed hash error")
	}
}

func testArgon2Parameters() Argon2Parameters {
	return Argon2Parameters{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
}
