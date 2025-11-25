package logging

import (
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
)

var (
	root *slog.Logger
	once sync.Once
)

type Config struct {
	AppName string
	Level   slog.Leveler
	Format  string // "json" or "text"
}

func Init(cfg Config) {
	once.Do(func() {
		var handler slog.Handler

		switch cfg.Format {
		case "json":
			handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: cfg.Level,
			})

		default:
			handler = tint.NewHandler(os.Stdout, &tint.Options{
				Level: cfg.Level,
			})
		}
		root = slog.New(handler).With(
			"app", cfg.AppName,
		)

	})
}

func Root() *slog.Logger {
	if root == nil {
		Init(Config{AppName: "unknown", Level: slog.LevelInfo, Format: "json"})
	}

	return root
}

func ForService(service string) *slog.Logger {
	return Root().With()
}
