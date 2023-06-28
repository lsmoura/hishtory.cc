package main

import (
	"context"
	"gihtub.com/lsmoura/hishtory_cc/internal/ctl"
	"log"
	"os"
)

func run(ctx context.Context, args ...string) error {
	home := os.Getenv("HOME")
	if home == "" {
		home = "."
	}

	rootCmd := ctl.NewRootCmd(home)

	if err := rootCmd.ParseAndRun(ctx, args); err != nil {
		log.Fatal("root.ParseAndRun", "error", err)
	}

	return nil
}

func main() {
	if err := run(context.Background(), os.Args[1:]...); err != nil {
		panic(err)
	}
}
