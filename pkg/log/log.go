package log

import (
	"context"
	"golang.org/x/exp/slog"
	"io"
	"os"
)

type Leveler interface {
	Level() Level
}

type Level slog.Level

var (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelError = Level(slog.LevelError)
)

func (l Level) Level() Level {
	return l
}

type Logger interface {
	Log(ctx context.Context, level Leveler, msg string, args ...any)
	DebugCtx(ctx context.Context, msg string, args ...any)
	InfoCtx(ctx context.Context, msg string, args ...any)
	ErrorCtx(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

func Default() Logger {
	return &internalLogger{slog.Default()}
}

func New(out io.Writer, level Leveler) Logger {
	if out == nil {
		out = os.Stdout
	}

	lvl := slog.LevelInfo
	if level != nil {
		lvl = slog.Level(level.Level())
	}

	return &internalLogger{slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: lvl}))}
}

func log(ctx context.Context, level Leveler, msg string, args ...any) {
	logger := FromContext(ctx)
	if logger == nil {
		return
	}

	logger.Log(ctx, level.Level(), msg, args...)
}

func WithCtx(ctx context.Context, args ...any) Logger {
	logger := FromContext(ctx)
	if logger == nil {
		return logger
	}

	return logger.With(args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	log(ctx, LevelDebug, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	log(ctx, LevelInfo, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	log(ctx, LevelError, msg, args...)
}
