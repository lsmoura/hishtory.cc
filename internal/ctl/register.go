package ctl

import (
	"context"
	"flag"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/client"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/xid"
)

type RegisterConfig struct {
	*RootConfig
}

type RegisterCmd struct {
	Conf *RegisterConfig

	rootCmd *RootCmd

	*ffcli.Command
}

func NewRegisterCmd(root *RootCmd) *RegisterCmd {
	conf := RegisterConfig{
		RootConfig: root.Conf,
	}
	cmd := RegisterCmd{
		Conf:    &conf,
		rootCmd: root,
	}
	fs := flag.NewFlagSet("hishtory register", flag.ExitOnError)
	cmd.RegisterFlags(fs)

	cmd.Command = &ffcli.Command{
		Name:       "register",
		ShortUsage: "hishtory register [user_id]",
		ShortHelp:  "Register this device",
		FlagSet:    fs,
		Exec:       cmd.Exec,
	}
	return &cmd
}

func (c *RegisterCmd) RegisterFlags(_ *flag.FlagSet) {
	// blank for now
}

func (c *RegisterCmd) Exec(ctx context.Context, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("expected up to one argument: user_id")
	}

	if err := c.rootCmd.init(); err != nil {
		return fmt.Errorf("cannot init: %w", err)
	}

	var shouldSave bool
	deviceID := c.rootCmd.Settings.DeviceID
	if deviceID == "" {
		deviceID = xid.New().String()
		c.rootCmd.Settings.DeviceID = deviceID
		shouldSave = true
	}

	userSecret := c.rootCmd.Settings.UserSecret
	if userSecret.IsEmpty() {
		var err error
		userSecret, err = client.NewUserSecret()
		if err != nil {
			return fmt.Errorf("cannot generate user secret: %w", err)
		}
		c.rootCmd.Settings.UserSecret = userSecret
		shouldSave = true
	}

	if shouldSave {
		if err := c.rootCmd.Settings.Save(c.Conf.RootConfig.RootPath); err != nil {
			return fmt.Errorf("cannot save settings: %w", err)
		}
	}

	cli, err := c.Conf.RootConfig.Client(userSecret, deviceID)
	if err != nil {
		return fmt.Errorf("cannot get client: %w", err)
	}

	if err := cli.Register(ctx); err != nil {
		return fmt.Errorf("cannot register: %w", err)
	}

	return nil
}
