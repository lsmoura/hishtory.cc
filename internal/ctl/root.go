package ctl

import (
	"context"
	"flag"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/client"
	"gihtub.com/lsmoura/hishtory_cc/internal/localdb"
	"github.com/peterbourgon/ff/v3/ffcli"
	"os"
)

type RootConfig struct {
	Host     string
	RootPath string

	dbLocation string
}

func (c RootConfig) Client(userSecret *client.UserSecret, deviceID string) (*client.Client, error) {
	if c.Host == "" {
		return nil, fmt.Errorf("host is required")
	}

	return client.New(c.Host, userSecret, deviceID), nil
}

type RootCmd struct {
	Conf     *RootConfig
	DB       *localdb.DB
	Client   *client.Client
	Settings Settings

	defaultRoot string
	isInit      bool

	*ffcli.Command
}

func NewRootCmd(rootPath string) *RootCmd {
	fs := flag.NewFlagSet("hishtory", flag.ExitOnError)

	cmd := RootCmd{
		Conf:        &RootConfig{},
		defaultRoot: rootPath,
	}
	cmd.Command = &ffcli.Command{
		Name:       "hishtory",
		ShortUsage: "hishtory [flags] <command>",
		FlagSet:    fs,
		Exec:       cmd.Exec,
	}
	cmd.RegisterFlags(fs)

	cmd.Command.Subcommands = []*ffcli.Command{
		NewRegisterCmd(&cmd).Command,
		NewSaveCmd(&cmd).Command,
		//NewListCmd(&cmd).Command,
		//NewSearchCmd(&cmd).Command,
		//NewAddMovieCmd(&cmd).Command,
		//NewListReleasesCmd(&cmd).Command,
		//NewManualDownloadCmd(&cmd).Command,
		//NewReleaseUpdateCmd(&cmd).Command,
		//NewRefreshCmd(&cmd).Command,
	}

	return &cmd
}

func (c *RootCmd) init() error {
	if c.isInit {
		return nil
	}
	c.isInit = true

	if c.Conf.RootPath == "" {
		c.Conf.RootPath = c.defaultRoot
	}
	if c.Conf.RootPath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get user home dir: %w", err)
		}
		c.Conf.RootPath = fmt.Sprintf("%s/.config/hishtory", homedir)
	}

	configFn := fmt.Sprintf("%s/config.json", c.Conf.RootPath)
	settings, err := ReadSettings(configFn)
	if err != nil {
		return fmt.Errorf("cannot read settings: %w", err)
	}

	c.Settings = settings

	// db
	if c.DB == nil {
		var dbLocation string
		if c.Conf.dbLocation != "" {
			dbLocation = c.Conf.dbLocation
		} else {
			dbLocation = fmt.Sprintf("%s/hishtory.db", c.Conf.RootPath)
		}

		database, err := localdb.NewWithSQLiteDSN(dbLocation)
		if err != nil {
			return fmt.Errorf("failed to open database %q: %w", dbLocation, err)
		}

		if err := database.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate the DB: %w", err)
		}

		c.DB = database
	}

	// client
	if c.Client == nil && c.Conf.Host != "" && !settings.UserSecret.IsEmpty() {
		if err := c.initClient(settings.UserSecret, settings.DeviceID); err != nil {
			return fmt.Errorf("cannot create client: %w", err)
		}
	}

	return nil
}

func (c *RootCmd) initClient(userSecret *client.UserSecret, deviceID string) error {
	if c.Client == nil && c.Conf.Host != "" && !userSecret.IsEmpty() {
		hClient, err := c.Conf.Client(userSecret, deviceID)
		if err != nil {
			return fmt.Errorf("cannot create client: %w", err)
		}

		c.Client = hClient
		c.Settings.DeviceID = deviceID
		c.Settings.UserSecret = userSecret
	}

	return nil
}

func (c *RootCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Conf.Host, "host", "localhost:8080", "hishtory server host")
	fs.StringVar(&c.Conf.RootPath, "root", "", "root path (defaults to HOME/.config/hishtory)")
}

func (c *RootCmd) Exec(_ context.Context, _ []string) error {
	c.FlagSet.Usage()
	return nil
}
