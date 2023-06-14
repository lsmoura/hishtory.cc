package db

import (
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/model"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func New(db *gorm.DB) *DB {
	return &DB{db: db}
}

func (db *DB) Migrate() error {
	//db.AutoMigrate(&shared.EncHistoryEntry{})
	//db.AutoMigrate(&shared.Device{})
	//db.AutoMigrate(&database.UsageData{})
	//db.AutoMigrate(&shared.DumpRequest{})
	//db.AutoMigrate(&shared.DeletionRequest{})
	//db.AutoMigrate(&shared.Feedback{})
	if err := db.db.AutoMigrate(&model.Device{}); err != nil {
		return fmt.Errorf("failed to migrate Device: %w", err)
	}

	return nil
}
