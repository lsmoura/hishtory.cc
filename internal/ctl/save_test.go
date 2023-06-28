package ctl

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSaveEmpty(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewSaveCmd(&RootCmd{
		Conf:        &RootConfig{},
		Settings:    Settings{},
		defaultRoot: tmp,
	})

	err := cmd.Exec(nil, []string{})
	require.Error(t, err, "cmd.Exec")
}

func TestSaveHappyPath(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewSaveCmd(&RootCmd{
		Conf: &RootConfig{
			dbLocation: ":memory:",
		},
		Settings: Settings{
			Hostname: "testhost",
			DeviceID: "testdevice",
		},
		defaultRoot: tmp,
	})

	err := cmd.Exec(nil, []string{"cat", "/etc/passwd"})
	require.NoError(t, err, "cmd.Exec")

	// make sure it was saved to DB
	entries, err := cmd.rootCmd.DB.Entries()
	require.NoError(t, err, "db.Entries")
	assert.Len(t, entries, 1, "db.Entries")
}
