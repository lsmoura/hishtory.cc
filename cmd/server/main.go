package main

import (
	"context"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/db"
	"gihtub.com/lsmoura/hishtory_cc/internal/server"
	"gihtub.com/lsmoura/hishtory_cc/pkg/log"
	"os"
)

func main() {
	logger := log.Default()
	ctx := log.WithContext(context.Background(), logger)

	log.InfoCtx(ctx, "starting hishtory server")

	if region := os.Getenv("FLY_REGION"); region == "" {
		log.UpdateContextWith(ctx, "region", region)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.NewWithPostgresDSN(os.Getenv("DB_URL"))
	if err != nil {
		log.ErrorCtx(context.Background(), "failed to connect to the DB", "err", err)
		return
	}

	if err := database.Migrate(); err != nil {
		log.ErrorCtx(context.Background(), "failed to migrate the DB", "err", err)
		return
	}

	addr := fmt.Sprintf(":%s", port)

	s := server.New(database)

	if err := s.Start(ctx, addr); err != nil {
		log.ErrorCtx(ctx, "server failed", "err", err)
	}
}
