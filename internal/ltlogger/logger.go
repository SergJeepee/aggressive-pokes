package ltlogger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(structured bool, app string, level slog.Level) Logger {
	loggerOpts := &slog.HandlerOptions{
		Level: level,
	}

	var logger *slog.Logger
	if structured {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, loggerOpts)).With("app", app)
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, loggerOpts)).With("app", app)
	slog.SetDefault(logger)
	return Logger{logger}
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args)
	os.Exit(1)
}
