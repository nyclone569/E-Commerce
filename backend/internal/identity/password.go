package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/text/unicode/norm"
)

const argon2Version = 19

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(encodedHash, password string) (bool, error)
}

type Argon2Parameters struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Argon2idHasher struct {
	parameters Argon2Parameters
}

func NewArgon2idHasher(parameters Argon2Parameters) *Argon2idHasher {
	return &Argon2idHasher{parameters: parameters}
}

func DefaultArgon2Parameters() Argon2Parameters {
	return Argon2Parameters{
		MemoryKiB:   19 * 1024,
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func (h *Argon2idHasher) Hash(password string) (string, error) {
	if err := validateArgon2Parameters(h.parameters); err != nil {
		return "", err
	}
	salt := make([]byte, h.parameters.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	password = norm.NFC.String(password)
	hash := argon2.IDKey([]byte(password), salt, h.parameters.Iterations, h.parameters.MemoryKiB, h.parameters.Parallelism, h.parameters.KeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2Version,
		h.parameters.MemoryKiB,
		h.parameters.Iterations,
		h.parameters.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *Argon2idHasher) Verify(encodedHash, password string) (bool, error) {
	parameters, salt, expectedHash, err := parseArgon2idHash(encodedHash)
	if err != nil {
		return false, err
	}
	password = norm.NFC.String(password)
	actualHash := argon2.IDKey([]byte(password), salt, parameters.Iterations, parameters.MemoryKiB, parameters.Parallelism, parameters.KeyLength)
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}

func parseArgon2idHash(encodedHash string) (Argon2Parameters, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return Argon2Parameters{}, nil, nil, errors.New("invalid Argon2id hash format")
	}

	var memory, iterations uint64
	var parallelism uint64
	for _, parameter := range strings.Split(parts[3], ",") {
		keyValue := strings.SplitN(parameter, "=", 2)
		if len(keyValue) != 2 {
			return Argon2Parameters{}, nil, nil, errors.New("invalid Argon2id parameters")
		}
		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return Argon2Parameters{}, nil, nil, errors.New("invalid Argon2id parameter value")
		}
		switch keyValue[0] {
		case "m":
			memory = value
		case "t":
			iterations = value
		case "p":
			parallelism = value
		default:
			return Argon2Parameters{}, nil, nil, errors.New("unknown Argon2id parameter")
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2Parameters{}, nil, nil, errors.New("invalid Argon2id salt")
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2Parameters{}, nil, nil, errors.New("invalid Argon2id hash")
	}
	if parallelism > 255 {
		return Argon2Parameters{}, nil, nil, errors.New("invalid Argon2id parallelism")
	}
	parameters := Argon2Parameters{
		MemoryKiB: uint32(memory), Iterations: uint32(iterations), Parallelism: uint8(parallelism),
		SaltLength: uint32(len(salt)), KeyLength: uint32(len(expectedHash)),
	}
	if err := validateArgon2Parameters(parameters); err != nil {
		return Argon2Parameters{}, nil, nil, err
	}
	return parameters, salt, expectedHash, nil
}

func validateArgon2Parameters(parameters Argon2Parameters) error {
	if parameters.MemoryKiB < 8*1024 || parameters.MemoryKiB > 1024*1024 {
		return errors.New("Argon2id memory must be between 8 MiB and 1 GiB")
	}
	if parameters.Iterations < 1 || parameters.Iterations > 10 {
		return errors.New("Argon2id iterations must be between 1 and 10")
	}
	if parameters.Parallelism < 1 || parameters.Parallelism > 16 {
		return errors.New("Argon2id parallelism must be between 1 and 16")
	}
	if parameters.SaltLength < 16 || parameters.SaltLength > 64 {
		return errors.New("Argon2id salt length must be between 16 and 64 bytes")
	}
	if parameters.KeyLength < 16 || parameters.KeyLength > 64 {
		return errors.New("Argon2id key length must be between 16 and 64 bytes")
	}
	return nil
}
