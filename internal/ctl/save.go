package ctl

import (
	"context"
	"flag"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/peterbourgon/ff/v3/ffcli"
	"os"
	"os/user"
	"strings"
	"time"
)

type SaveConfig struct {
	*RootConfig

	StartTime               int64
	Workdir                 string
	ExitCode                int
	CurrentWorkingDirectory string

	NoPush bool
}

type SaveCmd struct {
	Conf *SaveConfig

	rootCmd *RootCmd

	*ffcli.Command
}

func NewSaveCmd(root *RootCmd) *SaveCmd {
	conf := SaveConfig{
		RootConfig: root.Conf,
	}
	cmd := SaveCmd{
		Conf:    &conf,
		rootCmd: root,
	}
	fs := flag.NewFlagSet("hishtory save", flag.ExitOnError)
	cmd.RegisterFlags(fs)

	cmd.Command = &ffcli.Command{
		Name:       "save",
		ShortUsage: "hishtory save command",
		ShortHelp:  "Register this device",
		FlagSet:    fs,
		Exec:       cmd.Exec,
	}
	return &cmd
}

func (c *SaveCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.Int64Var(&c.Conf.StartTime, "start", 0, "start time")
	fs.StringVar(&c.Conf.Workdir, "workdir", "", "workdir")
	fs.IntVar(&c.Conf.ExitCode, "exitcode", 0, "exitcode")
	fs.StringVar(&c.Conf.CurrentWorkingDirectory, "cwd", "", "current working directory")
}

func (c *SaveCmd) Exec(ctx context.Context, args []string) error {
	if err := c.rootCmd.init(); err != nil {
		return fmt.Errorf("cannot init: %w", err)
	}

	if len(args) == 0 {
		return fmt.Errorf("save takes at least one argument")
	}

	command := strings.Join(args, " ")

	var startTime time.Time
	if c.Conf.StartTime != 0 {
		var err error
		startTime = time.Unix(c.Conf.StartTime, 0)
		if err != nil {
			return fmt.Errorf("cannot parse start time: %w", err)
		}
	} else {
		startTime = time.Now()
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("cannot get hostname: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get home directory: %w", err)
	}

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("cannot get current user: %w", err)
	}

	deviceID, err := c.rootCmd.Settings.GetDeviceID()
	if err != nil {
		return fmt.Errorf("cannot get device ID: %w", err)
	}

	entry := model.HistoryEntry{
		LocalUsername:           currentUser.Username,
		Hostname:                hostname,
		Command:                 command,
		CurrentWorkingDirectory: c.Conf.CurrentWorkingDirectory,
		HomeDirectory:           homeDir,
		ExitCode:                c.Conf.ExitCode,
		StartTime:               startTime,
		EndTime:                 time.Now(),
		DeviceID:                deviceID,
	}

	if err := c.rootCmd.DB.SaveEntry(&entry); err != nil {
		return fmt.Errorf("cannot save entry: %w", err)
	}

	return nil
}
