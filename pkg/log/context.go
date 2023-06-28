package log

import (
	"context"
)

type contextKey struct{}

var logContextKey contextKey

type ctxValue struct {
	logger Logger
}

// WithContext returns a new context with the given log context.
func WithContext(ctx context.Context, logContext Logger) context.Context {
	if logContext == nil {
		return ctx
	}
	v := ctxValue{logger: logContext}
	return context.WithValue(ctx, logContextKey, &v)
}

// FromContext returns the log context from the given context.
func FromContext(ctx context.Context) Logger {
	v, ok := ctx.Value(logContextKey).(*ctxValue)
	if !ok {
		return nil
	}
	return v.logger
}

func UpdateContextWith(ctx context.Context, args ...any) {
	v, ok := ctx.Value(logContextKey).(*ctxValue)
	if !ok {
		return
	}

	v.logger = v.logger.With(args...)

}
