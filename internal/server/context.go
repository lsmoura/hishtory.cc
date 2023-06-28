package server

import "context"

type contextKey struct{}

var remoteAddrContextKey contextKey

// WithRemoteAddr returns a new context with the given remote address.
func WithRemoteAddr(ctx context.Context, remoteAddr string) context.Context {
	return context.WithValue(ctx, remoteAddrContextKey, remoteAddr)
}

// RemoteAddrFromContext returns the remote address from the given context.
func RemoteAddrFromContext(ctx context.Context) string {
	remoteAddr, ok := ctx.Value(remoteAddrContextKey).(string)
	if !ok {
		return ""
	}
	return remoteAddr
}
