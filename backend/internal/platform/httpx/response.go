package httpx

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error Error `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("encode HTTP response", "error", err)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]string) {
	WriteJSON(w, status, errorEnvelope{Error: Error{Code: code, Message: message, Details: details}})
}

func DecodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, (1<<20)+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return &multipleJSONValuesError{}
	}
	return nil
}

type multipleJSONValuesError struct{}

func (*multipleJSONValuesError) Error() string { return "request body must contain one JSON object" }
