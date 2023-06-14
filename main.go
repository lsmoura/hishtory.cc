package main

import (
	"context"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/db"
	_ "github.com/lib/pq"
	"golang.org/x/exp/slog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
)

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoCtx(r.Context(), "request received", "method", r.Method, "url", r.URL)
		next.ServeHTTP(w, r)
	})
}

func start(addr string, region string) error {
	http.Handle("/", loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello from Fly! The current region is %s and the hostname is %s.\n", region, r.Host)
	})))

	slog.InfoCtx(context.Background(), "starting server", "addr", addr)
	return http.ListenAndServe(addr, nil)
}

func getDB() (*db.DB, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL not set")
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: dbURL}))
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	return db.New(gormDB), nil
}

func main() {
	var region string
	if region = os.Getenv("FLY_REGION"); region == "" {
		region = ""
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := getDB()
	if err != nil {
		slog.ErrorCtx(context.Background(), "failed to connect to the DB", "err", err)
		return
	}

	if err := db.Migrate(); err != nil {
		slog.ErrorCtx(context.Background(), "failed to migrate the DB", "err", err)
		return
	}

	addr := fmt.Sprintf(":%s", port)

	if err := start(addr, region); err != nil {
		slog.ErrorCtx(context.Background(), "server failed", "err", err)
	}
}
