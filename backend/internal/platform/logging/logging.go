package logging

import (
	"log/slog"
	"os"
)

func New(service, environment, version string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler).With(
		slog.String("service", service),
		slog.String("environment", environment),
		slog.String("version", version),
	)
}
