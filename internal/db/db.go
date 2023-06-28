package db

import (
	"context"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	db *gorm.DB

	isConcurrentSupported bool
}

var (
	ErrUserNotFound     = fmt.Errorf("user not found")
	ErrInvalidParameter = fmt.Errorf("invalid parameter")
)

func New(db *gorm.DB) *DB {
	return &DB{db: db}
}

func NewWithPostgresDSN(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn")
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}))
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	db := New(gormDB)
	db.isConcurrentSupported = true

	return db, nil
}

func NewWithSQLiteDSN(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn")
	}
	gormDB, err := gorm.Open(sqlite.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	return New(gormDB), nil
}

func (db *DB) Close() error {
	sqldb, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("cannot get db: %w", err)
	}

	return sqldb.Close()
}

func (db *DB) Ping() error {
	sqldb, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("cannot get db: %w", err)
	}

	return sqldb.Ping()
}

func (db *DB) createIndex(indexName, tableName string, columnNames []string) error {
	columnName := strings.Join(columnNames, ", ")

	if db.isConcurrentSupported {
		resp := db.db.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s USING btree(%s)", indexName, tableName, columnName))

		if resp.Error != nil {
			return fmt.Errorf("db.Exec (concurrently): %w", resp.Error)
		}
	}

	resp := db.db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s)", indexName, tableName, columnName))
	if resp.Error != nil {
		return fmt.Errorf("db.Exec: %w", resp.Error)
	}

	return nil
}

func (db *DB) Migrate() error {
	if err := db.db.AutoMigrate(&model.EncHistoryEntry{}); err != nil {
		return fmt.Errorf("failed to migrate EncHistoryEntry: %w", err)
	}

	encHistoryEntryIndexes := []struct {
		indexName   string
		tableName   string
		columnNames []string
	}{
		{"user_id_idx", "enc_history_entries", []string{"read_count"}},
		{"device_id_idx", "enc_history_entries", []string{"device_id"}},
		{"read_count_idx", "enc_history_entries", []string{"user_id", "device_id", "date"}},
	}
	for _, index := range encHistoryEntryIndexes {
		if err := db.createIndex(index.indexName, index.tableName, index.columnNames); err != nil {
			return fmt.Errorf("failed to create index %q: %w", index.indexName, err)
		}
	}

	//db.AutoMigrate(&database.UsageData{})
	//db.AutoMigrate(&shared.DumpRequest{})
	//db.AutoMigrate(&shared.DeletionRequest{})
	//db.AutoMigrate(&shared.Feedback{})
	if err := db.db.AutoMigrate(&model.Device{}); err != nil {
		return fmt.Errorf("failed to migrate Device: %w", err)
	}

	return nil
}

func (db *DB) RegisterDevice(ctx context.Context, userID, deviceID, remoteIP string) error {
	if userID == "" || deviceID == "" {
		return ErrInvalidParameter
	}
	resp := db.db.WithContext(ctx).Create(&model.Device{
		UserID:           userID,
		DeviceID:         deviceID,
		RegistrationIP:   remoteIP,
		RegistrationDate: time.Now(),
	})

	if resp.Error != nil {
		return fmt.Errorf("db.Create: %w", resp.Error)
	}

	return nil
}

func (db *DB) InsertHistoryEntries(ctx context.Context, entries []*model.EncHistoryEntry) error {
	if len(entries) == 0 {
		return nil
	}

	userID := entries[0].UserID

	if devices, err := db.DevicesForUser(ctx, userID); err != nil {
		return fmt.Errorf("db.DevicesForUser: %w", err)
	} else if len(devices) < 1 {
		return fmt.Errorf("found no devices associated with user_id=%s, can't save history entry: %w", userID, ErrUserNotFound)
	}

	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			if entry.UserID != userID {
				return fmt.Errorf("user_id mismatch %s != %s: %w", entry.UserID, userID, ErrInvalidParameter)
			}
			if entry.Date.IsZero() {
				entry.Date = time.Now()
			}
			resp := tx.Create(entry)

			if resp.Error != nil {
				return fmt.Errorf("tx.Create: %w", resp.Error)
			}
		}

		return nil
	})
}

func (db *DB) Entries(ctx context.Context) ([]*model.EncHistoryEntry, error) {
	var entries []*model.EncHistoryEntry
	resp := db.db.WithContext(ctx).Find(&entries)

	if resp.Error != nil {
		return nil, fmt.Errorf("db.Find: %w", resp.Error)
	}

	return entries, nil
}

func (db *DB) GetHistoryEntriesForUser(ctx context.Context, userID string) ([]*model.EncHistoryEntry, error) {
	var historyEntries []*model.EncHistoryEntry
	resp := db.db.WithContext(ctx).Where("user_id = ?", userID).Find(&historyEntries)

	if resp.Error != nil {
		return nil, fmt.Errorf("db.Where.Find: %w", resp.Error)
	}

	return historyEntries, nil
}

func (db *DB) Devices(ctx context.Context) ([]*model.Device, error) {
	var devices []*model.Device
	resp := db.db.WithContext(ctx).Find(&devices)

	if resp.Error != nil {
		return nil, fmt.Errorf("db.Find: %w", resp.Error)
	}

	return devices, nil
}

func (db *DB) DevicesForUser(ctx context.Context, userID string) ([]*model.Device, error) {
	var devices []*model.Device
	resp := db.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices)

	if resp.Error != nil {
		return nil, fmt.Errorf("db.Where.Find: %w", resp.Error)
	}

	return devices, nil
}
