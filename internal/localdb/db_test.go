package localdb

import (
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func testDB(t *testing.T) *DB {
	t.Helper()

	db, err := NewWithSQLiteDSN(":memory:")
	require.NoError(t, err, "NewWithSQLiteDSN")

	require.NoError(t, db.Migrate())

	return db
}

func TestMigration(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "db.Close")
	})

	require.NoError(t, db.Migrate(), "db.Migrate")
}

func TestSaveEntry(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "db.Close")
	})

	entry := &model.HistoryEntry{
		DeviceID: "foo",
		Command:  "bar",
	}

	require.NoError(t, db.SaveEntry(entry), "db.SaveEntry")

	entries, err := db.Entries()
	require.NoError(t, err, "db.Entries")
	assert.Len(t, entries, 1, "db.Entries")

	returnEntry := entries[0]
	assert.Equal(t, returnEntry.DeviceID, "foo")
	assert.Equal(t, returnEntry.Command, "bar")
}

func TestQuery(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "db.Close")
	})

	entries := []*model.HistoryEntry{
		{
			DeviceID:      "foo",
			Command:       `echo -n "foobar 123" | sha256sum`,
			LocalUsername: "root",
		},
		{
			DeviceID:      "bar",
			Command:       `pwd`,
			LocalUsername: "root",
		},
		{
			DeviceID:      "foo",
			Command:       `cat /etc/passwd`,
			ExitCode:      1,
			LocalUsername: "user",
		},
		{
			DeviceID:      "foo",
			Command:       `cat /etc/passwd`,
			ExitCode:      0,
			LocalUsername: "root",
		},
		{
			DeviceID:      "bar",
			Command:       `cat /etc/passwd`,
			ExitCode:      1,
			LocalUsername: "user",
		},
		{
			DeviceID:      "something",
			Command:       `cat /etc/passwd`,
			ExitCode:      1,
			LocalUsername: "user",
		},
	}
	for _, entry := range entries {
		require.NoErrorf(t, db.SaveEntry(entry), "db.SaveEntry: %+v", entry)
	}

	queries := []struct {
		name     string
		query    string
		expected []*model.HistoryEntry
	}{
		{
			name:     "empty query",
			query:    "",
			expected: entries,
		},
		{
			name:  "pwd query",
			query: "pwd",
			expected: []*model.HistoryEntry{
				entries[1],
			},
		},
		{
			name:  "sha256sum query",
			query: "sha256sum",
			expected: []*model.HistoryEntry{
				entries[0],
			},
		},
		{
			name:     "hosts empty result with no query",
			query:    "host:xyz",
			expected: []*model.HistoryEntry{},
		},
		{
			name:     "hosts empty result with query I",
			query:    "pwd host:foo",
			expected: []*model.HistoryEntry{},
		},
		{
			name:     "hosts empty result with query II",
			query:    "host:foo pwd",
			expected: []*model.HistoryEntry{},
		},
		{
			name:  "hosts result with query",
			query: "host:bar pwd",
			expected: []*model.HistoryEntry{
				entries[1],
			},
		},
		{
			name:     "multi hosts result with query",
			query:    "host:bar cat host:foo",
			expected: []*model.HistoryEntry{entries[2], entries[3], entries[4]},
		},
		{
			name:     "user name and exit code query",
			query:    "user:root exit_code:0 /etc/passwd",
			expected: []*model.HistoryEntry{entries[3]},
		},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			returnedEntries, err := db.Query(tt.query)
			require.NoError(t, err, "db.Query")

			assert.Equal(t, tt.expected, returnedEntries)
		})
	}
}

func TestUpdateEntry(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "db.Close")
	})

	entry := model.HistoryEntry{
		LocalUsername:           "root",
		Hostname:                "localhost",
		Command:                 "cat /etc/passwd",
		CurrentWorkingDirectory: "/root",
		HomeDirectory:           "/root",
		ExitCode:                0,
		StartTime:               time.Now().Add(-time.Second),
		EndTime:                 time.Now(),
		Pushed:                  false,
	}

	require.NoError(t, db.SaveEntry(&entry), "db.SaveEntry")

	entry.Pushed = true

	require.NoError(t, db.UpdateEntry(&entry), "db.UpdateEntry")

	entries, err := db.Entries()
	require.NoError(t, err, "db.Entries")

	assert.Len(t, entries, 1, "db.Entries")
	assert.Equal(t, entries[0].Pushed, true)
}
