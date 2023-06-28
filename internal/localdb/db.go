package localdb

import (
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type DB struct {
	db *gorm.DB
}

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
	if gormDB == nil {
		return nil, fmt.Errorf("gorm.Open: nil")
	}

	return New(gormDB), nil
}

func NewWithSQLiteDSN(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn")
	}
	gormDB, err := gorm.Open(sqlite.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}
	if gormDB == nil {
		return nil, fmt.Errorf("gorm.Open: nil")
	}

	return New(gormDB), nil
}

func (d *DB) Migrate() error {
	if err := d.db.AutoMigrate(&model.HistoryEntry{}); err != nil {
		return fmt.Errorf("AutoMigrate HistoryEntry: %w", err)
	}

	return nil
}

func (d *DB) Entries() ([]*model.HistoryEntry, error) {
	var entries []*model.HistoryEntry
	tx := d.db.Model(&model.HistoryEntry{}).Find(&entries)
	if tx.Error != nil {
		return nil, fmt.Errorf("Find: %w", tx.Error)
	}

	return entries, nil
}

func (d *DB) UnsyncdEntries() ([]*model.HistoryEntry, error) {
	var entries []*model.HistoryEntry
	tx := d.db.Model(&model.HistoryEntry{}).Where("pushed IS FALSE").Find(&entries)
	if tx.Error != nil {
		return nil, fmt.Errorf("Find: %w", tx.Error)
	}

	return entries, nil
}

func (d *DB) SaveEntry(entry *model.HistoryEntry) error {
	if d == nil {
		return fmt.Errorf("database was not initialized")
	}
	if d.db == nil {
		return fmt.Errorf("db is nil")
	}
	tx := d.db.Create(entry)

	if tx.Error != nil {
		return fmt.Errorf("Create: %w", tx.Error)
	}

	return nil
}

func (d *DB) UpdateEntry(entry *model.HistoryEntry) error {
	tx := d.db.Model(entry).
		Updates(map[string]any{"pushed": entry.Pushed, "pulled": entry.Pulled})

	if tx.Error != nil {
		return fmt.Errorf("Save: %w", tx.Error)
	}

	return nil
}

// Query returns all the entries that matches the query.
// special cases:
// - if query is empty, returns all entries.
// - a `host:` parameter will filter queries that contains the host value by deviceID
// - a `exit_code:` parameter will filter queries that contains the exit code value by exitCode
// - a `user:` parameter will filter queries that contains the user value by userName
// if multiple entries of the same special query parameter (like "host:") are provided, they will be ORed.
func (d *DB) Query(query string) ([]*model.HistoryEntry, error) {
	if query == "" {
		return d.Entries()
	}

	// TODO: respect quotes
	queryPieces := strings.Split(query, " ")
	var hosts []string
	var exitCodes []int64
	var userNames []string
	var resultQueryPieces []string
	for _, piece := range queryPieces {
		if strings.HasPrefix(piece, "host:") {
			hosts = append(hosts, strings.TrimPrefix(piece, "host:"))
			continue
		}
		if strings.HasPrefix(piece, "exit_code:") {
			exitCode := strings.TrimPrefix(piece, "exit_code:")
			v, err := strconv.ParseInt(exitCode, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid exit code value %q: %w", exitCode, err)
			}
			exitCodes = append(exitCodes, v)
			continue
		}
		if strings.HasPrefix(piece, "user:") {
			userNames = append(userNames, strings.TrimPrefix(piece, "user:"))
			continue
		}
		resultQueryPieces = append(resultQueryPieces, piece)
	}

	var entries []*model.HistoryEntry
	actualQuery := strings.Join(resultQueryPieces, " ")
	tx := d.db.Model(&model.HistoryEntry{})
	if len(hosts) > 0 {
		var hostsWhereClauses []string
		for _, host := range hosts {
			hostsWhereClauses = append(hostsWhereClauses, "device_id LIKE '%"+host+"%'")
		}
		tx = tx.Where(strings.Join(hostsWhereClauses, " OR "))
	}
	if len(exitCodes) > 0 {
		var exitCodesWhereClauses []string
		for _, exitCode := range exitCodes {
			exitCodesWhereClauses = append(exitCodesWhereClauses, "exit_code = "+strconv.FormatInt(int64(exitCode), 10))
		}
		tx = tx.Where(strings.Join(exitCodesWhereClauses, " OR "))
	}
	if len(userNames) > 0 {
		var userNamesWhereClauses []string
		for _, userName := range userNames {
			userNamesWhereClauses = append(userNamesWhereClauses, "local_username LIKE '%"+userName+"%'")
		}
		tx = tx.Where(strings.Join(userNamesWhereClauses, " OR "))
	}
	if actualQuery != "" {
		tx = tx.Where("command LIKE ?", "%"+actualQuery+"%")
	}
	tx = tx.Find(&entries)
	if tx.Error != nil {
		return nil, fmt.Errorf("tx.Where.Find: %w", tx.Error)
	}

	return entries, nil
}

func (d *DB) Close() error {
	sqldb, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("cannot get db: %w", err)
	}

	return sqldb.Close()
}
