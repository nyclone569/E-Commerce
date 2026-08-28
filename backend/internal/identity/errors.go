package identity

import "errors"

var (
	ErrConflict           = errors.New("identity conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("authentication required")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "identity validation failed"
}
