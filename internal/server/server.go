package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"gihtub.com/lsmoura/hishtory_cc/internal/db"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"gihtub.com/lsmoura/hishtory_cc/pkg/log"
)

type Server struct {
	db *db.DB

	server *http.Server
}

func New(db *db.DB) *Server {
	return &Server{db: db}
}

func (s *Server) mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/submit", s.handleSubmit())
	mux.Handle("/api/v1/query", s.handleQuery())
	mux.Handle("/api/v1/register", s.handleRegister())

	mux.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Expires", "Thu, 31 Dec 2037 23:55:55 GMT")
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.Handle("/health", s.handleHealth())
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "hishtory.cc")
	}))

	return mux
}

func (s *Server) baseListener(ctx context.Context) func(net.Listener) context.Context {
	return func(_ net.Listener) context.Context {
		rCtx := ctx
		if rCtx == nil {
			rCtx = context.Background()
		}

		if logger := log.FromContext(rCtx); logger != nil {
			rCtx = log.WithContext(rCtx, logger.With())
		}

		return rCtx
	}
}

func (s *Server) Start(ctx context.Context, addr string) error {
	log.InfoCtx(ctx, "starting server", "addr", addr)

	mux := s.mux()

	s.server = &http.Server{
		Addr:        addr,
		Handler:     applyMiddlewares(mux, panicGuardMiddleware, correlationMiddleware, extractRemoteAddr, loggerMiddleware),
		BaseContext: s.baseListener(ctx),
	}

	return s.server.ListenAndServe()
}

func (s *Server) handleHealth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.db.Ping(); err != nil {
			log.ErrorCtx(r.Context(), "failed to ping DB", "error", err)
			http.Error(w, "failed to ping DB", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// handleSubmit handles the submission of history entries, POST only
func (s *Server) handleSubmit() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.InfoCtx(r.Context(), "handleSubmit")
		if r.Method != http.MethodPost {
			log.ErrorCtx(r.Context(), "invalid method", "method", r.Method)
			http.Error(w, "invalid method", http.StatusBadRequest)
			return
		}

		var entries []*model.EncHistoryEntry
		if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
			log.ErrorCtx(r.Context(), "failed to decode request body", "error", err)
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}

		if err := s.db.InsertHistoryEntries(r.Context(), entries); err != nil {
			if errors.Is(err, db.ErrUserNotFound) || errors.Is(err, db.ErrInvalidParameter) {
				log.InfoCtx(r.Context(), "failed to insert history entries", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.ErrorCtx(r.Context(), "failed to insert history entries", "error", err)
			http.Error(w, "failed to insert history entries", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// handleQuery handles the query of history entries for a given user, GET only
func (s *Server) handleQuery() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.ErrorCtx(r.Context(), "invalid method", "method", r.Method)
			http.Error(w, "invalid method", http.StatusBadRequest)
			return
		}

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}

		entries, err := s.db.GetHistoryEntriesForUser(r.Context(), userID)
		if err != nil {
			log.ErrorCtx(r.Context(), "failed to get history entries", "error", err)
			http.Error(w, "failed to get history entries", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(entries); err != nil {
			log.ErrorCtx(r.Context(), "failed to encode response body", "error", err)
			http.Error(w, "failed to encode response body", http.StatusInternalServerError)
			return
		}
	})
}

// handleRegister handles the registration of a new device, GET only
func (s *Server) handleRegister() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodGet {
			log.ErrorCtx(r.Context(), "invalid method", "method", r.Method)
			http.Error(w, "invalid method", http.StatusBadRequest)
			return
		}

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}
		deviceID := r.URL.Query().Get("device_id")
		if deviceID == "" {
			http.Error(w, "missing device_id", http.StatusBadRequest)
			return
		}

		if err := s.db.RegisterDevice(ctx, userID, deviceID, RemoteAddrFromContext(ctx)); err != nil {
			log.ErrorCtx(ctx, "failed to register device", "device_id", deviceID, "user_id", userID, "error", err)
			http.Error(w, "failed to register device", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func (s *Server) Close() error {
	return s.server.Close()
}
