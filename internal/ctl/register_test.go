package ctl

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegisterHappyPath(t *testing.T) {
	serverDB, rootCmd := helperInitServerAndClient(t, nil)
	hClient := rootCmd.Client
	userSecret := rootCmd.Settings.UserSecret
	deviceID := rootCmd.Settings.DeviceID

	require.NoError(t, hClient.Register(context.Background()), "hClient.Register")

	// check devices
	devices, err := serverDB.Devices(context.Background())
	require.NoError(t, err, "serverDB.Devices")
	require.Len(t, devices, 1, "serverDB.Devices")
	assert.Equal(t, devices[0].UserID, userSecret.UserId(), "devices[0].UserID")
	assert.Equal(t, devices[0].DeviceID, deviceID, "devices[0].DeviceID")
}
