package server

import (
	log2 "gihtub.com/lsmoura/hishtory_cc/pkg/log"
	"github.com/rs/xid"
	"net/http"
	"time"
)

func applyMiddlewares(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	resp := h
	for i := len(middlewares) - 1; i >= 0; i-- {
		resp = middlewares[i](resp)
	}

	return resp
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			elapsed := time.Since(start)
			ipAddr := RemoteAddrFromContext(r.Context())
			log2.InfoCtx(r.Context(), "request", "method", r.Method, "url", r.URL, "elapsed", elapsed, "ip_addr", ipAddr)
		}()
		next.ServeHTTP(w, r)
	})
}

func panicGuardMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log2.ErrorCtx(r.Context(), "panic", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func extractRemoteAddr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		addr, ok := r.Header["X-Real-Ip"]
		if ok && len(addr) > 0 {
			ctx = WithRemoteAddr(ctx, addr[0])
		} else {
			ctx = WithRemoteAddr(ctx, r.RemoteAddr)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func correlationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		correlationID := xid.New().String()

		rw.Header().Add("X-Correlation-ID", correlationID)
		logger := log2.WithCtx(r.Context(), "correlation_id", correlationID)

		next.ServeHTTP(rw, r.WithContext(log2.WithContext(r.Context(), logger)))
	})
}
