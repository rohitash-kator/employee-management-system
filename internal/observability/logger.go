package observability

import (
	"log/slog"
	"os"
)

func NewLogger(appEnv string) *slog.Logger {
	level := slog.LevelInfo
	if appEnv == "local" || appEnv == "dev" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: appEnv == "local",
	})

	return slog.New(handler)
}
