package model

import (
	"database/sql/driver"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func TestHistoryEntryMigration(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err, "gorm.Open")

	err = gormDB.AutoMigrate(&HistoryEntry{})
	require.NoError(t, err, "gormDB.AutoMigrate")
}

func TestCustomColumns_Value(t *testing.T) {
	tests := []struct {
		name    string
		columns CustomColumns
		want    driver.Value
	}{
		{
			"regular",
			CustomColumns{
				{"foo", "bar"},
				{"baz", "qux"},
			},
			[]uint8(`[{"name":"baz","value":"qux"},{"name":"foo","value":"bar"}]`),
		},
		{
			"empty",
			CustomColumns{},
			[]uint8(`[]`),
		},
		{
			"double name",
			CustomColumns{
				{"foo", "bar"},
				{"foo", "zzz"},
				{"baz", "qux"},
			},
			[]uint8(`[{"name":"baz","value":"qux"},{"name":"foo","value":"bar"},{"name":"foo","value":"zzz"}]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.columns.Value()
			require.NoError(t, err, "Value()")
			assert.Equal(t, tt.want, v)
		})
	}
}

func TestCustomColumns_Scan(t *testing.T) {
	tests := []struct {
		name      string
		entry     any
		want      CustomColumns
		shouldErr bool
	}{
		{
			"regular []uint8",
			[]uint8(`[{"name":"baz","value":"qux"},{"name":"foo","value":"bar"}]`),
			CustomColumns{
				{"baz", "qux"},
				{"foo", "bar"},
			},
			false,
		},
		{
			"regular string",
			`[{"name":"baz","value":"qux"},{"name":"foo","value":"bar"}]`,
			CustomColumns{
				{"baz", "qux"},
				{"foo", "bar"},
			},
			false,
		},
		{
			"empty string",
			"",
			CustomColumns{},
			false,
		},
		{
			"nil",
			nil,
			CustomColumns{},
			false,
		},
		{
			"invalid json",
			`{"name":"baz","value":"qux"},{"name":"foo","value":"bar"}`,
			CustomColumns{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var columns CustomColumns
			err := columns.Scan(tt.entry)
			if tt.shouldErr {
				require.Error(t, err, "Scan()")
				return
			}
			require.NoError(t, err, "Scan()")
			assert.Equal(t, tt.want, columns)
		})
	}
}
