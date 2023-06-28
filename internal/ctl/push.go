package ctl

import (
	"context"
	"flag"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type PushConfig struct {
	*RootConfig

	Everything bool
}

type PushCmd struct {
	Conf *PushConfig

	rootCmd *RootCmd

	*ffcli.Command
}

func NewPushCmd(root *RootCmd) *PushCmd {
	conf := PushConfig{
		RootConfig: root.Conf,
	}
	cmd := PushCmd{
		Conf:    &conf,
		rootCmd: root,
	}
	fs := flag.NewFlagSet("hishtory push", flag.ExitOnError)
	cmd.RegisterFlags(fs)

	cmd.Command = &ffcli.Command{
		Name:       "save",
		ShortUsage: "hishtory push",
		ShortHelp:  "Push history to server",
		FlagSet:    fs,
		Exec:       cmd.Exec,
	}
	return &cmd
}

func (c *PushCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Conf.Everything, "everything", false, "push everything, regardless if it was already pushed")
}

func filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func (c *PushCmd) Exec(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("push takes no arguments")
	}

	if err := c.rootCmd.init(); err != nil {
		return fmt.Errorf("cannot init: %w", err)
	}

	db := c.rootCmd.DB
	resp, err := db.Entries()
	if err != nil {
		return fmt.Errorf("cannot get entries: %w", err)
	}

	if !c.Conf.Everything {
		resp = filter(resp, func(entry *model.HistoryEntry) bool {
			return entry.Pushed == false
		})
	}

	client := c.rootCmd.Client
	if client == nil {
		return fmt.Errorf("cannot push: no client")
	}

	if err := client.Save(ctx, resp); err != nil {
		return fmt.Errorf("cannot save: %w", err)
	}

	for _, entry := range resp {
		entry.Pushed = true
		if err := db.UpdateEntry(entry); err != nil {
			return fmt.Errorf("cannot update entry: %w", err)
		}
	}

	return nil
}
