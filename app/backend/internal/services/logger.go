package services

import (
	"log/slog"
	"os"

	"github.com/mephalrith/noodles/backend/internal/config"
)

var Logger *slog.Logger

func InitLogger(cfg *config.Config) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{}

	if cfg.IsProduction {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler).With("service", "noodles-dashboard")
}
