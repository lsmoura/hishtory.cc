package log

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"strings"
	"testing"
)

func TestLogContext_Empty(t *testing.T) {
	ctx := context.Background()
	logContext := WithContext(ctx, nil)

	require.Equal(t, ctx, logContext)
}

func TestLogContext_NoLogger(t *testing.T) {
	logger := FromContext(context.Background())
	assert.Nil(t, logger)
}

func TestLogContext_WithLogger(t *testing.T) {
	var buf strings.Builder

	baseLogger := slog.New(slog.NewTextHandler(&buf, nil))
	logger := &internalLogger{baseLogger}

	logContext := WithContext(context.Background(), logger)

	InfoCtx(logContext, "test")

	assert.Contains(t, buf.String(), "level=INFO")
	assert.Contains(t, buf.String(), "msg=test")
}

func TestLogContext_UpdateLogger(t *testing.T) {
	var buf strings.Builder

	baseLogger := slog.New(slog.NewTextHandler(&buf, nil))
	logger := &internalLogger{baseLogger}

	logContext := WithContext(context.Background(), logger)

	UpdateContextWith(logContext, "foo", "bar")

	InfoCtx(logContext, "test")
	logger.InfoCtx(context.Background(), "test2")

	assert.Contains(t, buf.String(), "level=INFO msg=test foo=bar\n")
	assert.Contains(t, buf.String(), "level=INFO msg=test2\n")
}

func TestLogContext_UpdateEmptyLogger(t *testing.T) {
	var buf strings.Builder

	baseLogger := slog.New(slog.NewTextHandler(&buf, nil))
	logger := &internalLogger{baseLogger}

	logContext := context.Background()

	UpdateContextWith(context.Background(), "foo", "bar")

	InfoCtx(logContext, "test")
	logger.InfoCtx(context.Background(), "test2")

	assert.NotContains(t, buf.String(), "foo=bar")
	assert.Contains(t, buf.String(), "level=INFO msg=test2\n")
}
