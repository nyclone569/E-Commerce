package catalog

import "errors"

var ErrConflict = errors.New("catalog resource already exists")
var ErrNotFound = errors.New("catalog resource not found")

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "catalog input is invalid" }
