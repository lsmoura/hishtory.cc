package log

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"strings"
	"testing"
)

func TestWithCtx(t *testing.T) {
	var buf strings.Builder

	baseLogger := slog.New(slog.NewTextHandler(&buf, nil))
	logger := &internalLogger{baseLogger}

	ctx := WithContext(context.Background(), logger)

	InfoCtx(ctx, "test")

	updatedLogger := WithCtx(ctx, "foo", "bar")

	updatedLogger.InfoCtx(ctx, "test2")

	// makes sure context logger was not updated
	InfoCtx(ctx, "other_test")

	assert.Contains(t, buf.String(), "level=INFO msg=test\n")
	assert.Contains(t, buf.String(), "level=INFO msg=test2 foo=bar\n")
	assert.Contains(t, buf.String(), "level=INFO msg=other_test\n")
}

func TestDebugCtx(t *testing.T) {
	var buf strings.Builder
	logger := New(&buf, LevelDebug)

	ctx := WithContext(context.Background(), logger)

	DebugCtx(ctx, "test")

	assert.Contains(t, buf.String(), "level=DEBUG")
	assert.Contains(t, buf.String(), "msg=test")
}

func TestInfoCtx(t *testing.T) {
	var buf strings.Builder
	logger := New(&buf, LevelDebug)

	ctx := WithContext(context.Background(), logger)

	InfoCtx(ctx, "test")

	assert.Contains(t, buf.String(), "level=INFO")
	assert.Contains(t, buf.String(), "msg=test")
}

func TestErrorCtx(t *testing.T) {
	var buf strings.Builder
	logger := New(&buf, LevelDebug)

	ctx := WithContext(context.Background(), logger)

	ErrorCtx(ctx, "test")

	assert.Contains(t, buf.String(), "level=ERROR")
	assert.Contains(t, buf.String(), "msg=test")
}
