package db

import (
	"context"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"testing"
)

func testDB(t *testing.T) *DB {
	t.Helper()

	db, err := NewWithSQLiteDSN(":memory:")
	require.NoError(t, err, "NewWithSQLiteDSN")
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "db.Close")
	})

	require.NoError(t, db.Migrate())

	return db
}

func TestRegisterDeviceFlow(t *testing.T) {
	db := testDB(t)

	userID := "foo"
	deviceID := "bar"

	require.NoError(t, db.RegisterDevice(context.Background(), userID, deviceID, ""), "RegisterDevice")

	t.Run("DevicesForUser with correct UserID", func(t *testing.T) {
		devicesForUser, err := db.DevicesForUser(context.Background(), userID)
		require.NoError(t, err, "DevicesForUser")

		require.Len(t, devicesForUser, 1)
	})

	t.Run("DevicesForUser with incorrect UserID", func(t *testing.T) {
		devicesForUser, err := db.DevicesForUser(context.Background(), userID+userID)
		require.NoError(t, err, "DevicesForUser")

		require.Len(t, devicesForUser, 0)
	})
}

func TestSubmitHistoryFlow(t *testing.T) {
	db := testDB(t)

	userID := "foo"
	deviceID := "bar"

	require.NoError(t, db.RegisterDevice(context.Background(), userID, deviceID, ""), "RegisterDevice")

	t.Run("SubmitHistory with correct UserID", func(t *testing.T) {
		entries := []*model.EncHistoryEntry{
			{
				UserID:        userID,
				DeviceID:      deviceID,
				EncryptedData: []byte("foo"),
			},
		}
		require.NoError(t, db.InsertHistoryEntries(context.Background(), entries), "SubmitHistory")
	})

	t.Run("SubmitHistory with invalid UserID", func(t *testing.T) {
		entries := []*model.EncHistoryEntry{
			{
				UserID:        userID + userID,
				DeviceID:      deviceID,
				EncryptedData: []byte("foo"),
			},
		}
		require.Error(t, db.InsertHistoryEntries(context.Background(), entries), "SubmitHistory")
	})

	t.Run("raise error when userIDs differ", func(t *testing.T) {
		entries := []*model.EncHistoryEntry{
			{
				UserID:        userID,
				DeviceID:      deviceID,
				EncryptedData: []byte("foo"),
			},
			{
				UserID:        userID + userID,
				DeviceID:      deviceID,
				EncryptedData: []byte("foo"),
			},
		}
		require.Error(t, db.InsertHistoryEntries(context.Background(), entries), "SubmitHistory")
	})

	t.Run("raise error when userID is empty", func(t *testing.T) {
		entries := []*model.EncHistoryEntry{
			{
				UserID:        "",
				DeviceID:      deviceID,
				EncryptedData: []byte("foo"),
			},
		}
		require.Error(t, db.InsertHistoryEntries(context.Background(), entries), "SubmitHistory")
	})
}

func TestSaveLoadEntries(t *testing.T) {
	userID := xid.New().String()
	deviceID := xid.New().String()

	entries := []*model.EncHistoryEntry{
		{
			UserID:        userID,
			DeviceID:      deviceID,
			EncryptedData: []byte("foo"),
		},
		{
			UserID:        userID,
			DeviceID:      deviceID,
			EncryptedData: []byte("bar"),
		},
	}

	db := testDB(t)

	require.NoError(t, db.RegisterDevice(context.Background(), userID, deviceID, "localhost"), "RegisterDevice")
	require.NoError(t, db.InsertHistoryEntries(context.Background(), entries), "db.InsertHistoryEntries")

	// fetch entries
	loadedEntries, err := db.Entries(context.Background())
	require.NoError(t, err, "db.LoadHistoryEntries")
	require.Len(t, loadedEntries, len(entries))

	// add more entries
	newUserID := xid.New().String()
	require.NoError(t, db.RegisterDevice(context.Background(), newUserID, deviceID, "localhost"), "RegisterDevice")
	require.NoError(t, db.InsertHistoryEntries(context.Background(), []*model.EncHistoryEntry{{
		UserID:        newUserID,
		DeviceID:      deviceID,
		EncryptedData: []byte("baz"),
	}}), "db.InsertHistoryEntries")
	loadedEntries, err = db.Entries(context.Background())
	require.NoError(t, err, "db.LoadHistoryEntries")
	require.Len(t, loadedEntries, len(entries)+1)

	// check the user filtering is working
	loadedEntries, err = db.GetHistoryEntriesForUser(context.Background(), userID)
	require.NoError(t, err, "db.LoadHistoryEntries")
	require.Len(t, loadedEntries, len(entries))
}
