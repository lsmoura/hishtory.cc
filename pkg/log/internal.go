package log

import (
	"context"
	"golang.org/x/exp/slog"
)

type internalLogger struct {
	*slog.Logger
}

func (logger *internalLogger) Log(ctx context.Context, level Leveler, msg string, args ...any) {
	if logger == nil {
		return
	}

	lvl := slog.LevelInfo
	if level != nil {
		lvl = slog.Level(level.Level())
	}

	logger.Logger.Log(ctx, lvl, msg, args...)
}

func (logger *internalLogger) With(args ...any) Logger {
	if logger == nil {
		return logger
	}
	resp := logger.Logger.With(args...)

	return &internalLogger{resp}
}
