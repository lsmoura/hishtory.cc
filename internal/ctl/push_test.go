package ctl

import (
	"context"
	"errors"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/client"
	"gihtub.com/lsmoura/hishtory_cc/internal/db"
	"gihtub.com/lsmoura/hishtory_cc/internal/server"
	"gihtub.com/lsmoura/hishtory_cc/pkg/log"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func helperInitServerAndClient(t *testing.T, logger log.Logger) (*db.DB, *RootCmd) {
	t.Helper()

	serverDB, err := db.NewWithSQLiteDSN(":memory:")
	require.NoError(t, err, "db.NewWithSQLiteDSN")
	require.NoError(t, serverDB.Migrate(), "serverDB.Migrate")
	hServer := server.New(serverDB)
	t.Cleanup(func() {
		require.NoError(t, hServer.Close(), "hServer.Close")
		require.NoError(t, serverDB.Close(), "serverDB.Close")
	})

	addr := ":8080"

	go func() {
		logCtx := context.Background()
		if logger != nil {
			logCtx = log.WithContext(context.Background(), logger)
		}
		err := hServer.Start(logCtx, addr)
		if !errors.Is(err, http.ErrServerClosed) {
			t.Error(fmt.Errorf("hServer.Start: %w", err))
		}
	}()

	userSecret, err := client.NewUserSecret()
	require.NoError(t, err, "client.NewUserSecret")

	deviceID := xid.New().String()
	tmp := t.TempDir()
	rootCmd := RootCmd{
		Conf: &RootConfig{
			dbLocation: ":memory:",
			Host:       fmt.Sprintf("http://localhost%s", addr),
		},
		Settings: Settings{
			Hostname:   "testhost",
			DeviceID:   deviceID,
			UserSecret: userSecret,
		},
		defaultRoot: tmp,
	}

	require.NoError(t, rootCmd.init(), "rootCmd.init")
	rootCmd.Settings.Hostname = "testhost"
	rootCmd.Settings.DeviceID = deviceID
	rootCmd.Settings.UserSecret = userSecret
	require.NoError(t, rootCmd.initClient(rootCmd.Settings.UserSecret, rootCmd.Settings.DeviceID), "rootCmd.initClient")

	return serverDB, &rootCmd
}

func TestPushHappyPath(t *testing.T) {
	serverDB, rootCmd := helperInitServerAndClient(t, nil)
	hClient := rootCmd.Client
	userSecret := rootCmd.Settings.UserSecret
	deviceID := rootCmd.Settings.DeviceID

	// register client
	require.NoError(t, hClient.Register(context.Background()), "hClient.Register")

	// check devices
	devices, err := serverDB.Devices(context.Background())
	require.NoError(t, err, "serverDB.Devices")
	require.Len(t, devices, 1, "serverDB.Devices")
	assert.Equal(t, devices[0].UserID, userSecret.UserId(), "devices[0].UserID")
	assert.Equal(t, devices[0].DeviceID, deviceID, "devices[0].DeviceID")

	saveCmd := NewSaveCmd(rootCmd)
	saveCmd.Conf.NoPush = true
	require.NoError(t, saveCmd.Exec(nil, []string{"cat", "/etc/passwd"}), "saveCmd.Exec")

	// make sure it was saved to DB
	entries, err := saveCmd.rootCmd.DB.Entries()
	require.NoError(t, err, "db.Entries")
	assert.Len(t, entries, 1, "db.Entries")

	pushCmd := NewPushCmd(rootCmd)
	require.NoError(t, pushCmd.Exec(nil, []string{}), "cmd.Exec")

	// check that the local entry was updated
	entries, err = saveCmd.rootCmd.DB.Entries()
	require.NoError(t, err, "db.Entries")
	require.Len(t, entries, 1, "db.Entries")
	assert.Truef(t, entries[0].Pushed, "entries[0].Pushed")

	// check that the server database has the correct entry
	serverEntries, err := serverDB.Entries(context.Background())
	require.NoError(t, err, "serverDB.Entries")
	require.Len(t, serverEntries, 1, "serverDB.Entries")
	assert.Equal(t, serverEntries[0].UserID, userSecret.UserId(), "serverEntries[0].UserID")
	assert.Equal(t, serverEntries[0].DeviceID, deviceID, "serverEntries[0].DeviceID")
}
