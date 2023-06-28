package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type HistoryEntry struct {
	LocalUsername           string    `json:"local_username" gorm:"primaryKey"`
	Hostname                string    `json:"hostname" gorm:"primaryKey"`
	Command                 string    `json:"command" gorm:"primaryKey"`
	CurrentWorkingDirectory string    `json:"current_working_directory" gorm:"primaryKey"`
	HomeDirectory           string    `json:"home_directory" gorm:"primaryKey"`
	ExitCode                int       `json:"exit_code" gorm:"primaryKey;autoIncrement:false"`
	StartTime               time.Time `json:"start_time" gorm:"primaryKey"`
	EndTime                 time.Time `json:"end_time" gorm:"primaryKey"`
	DeviceID                string    `json:"device_id" gorm:"primaryKey"`

	CustomColumns CustomColumns `json:"custom_columns,omitempty"`

	Pushed bool `json:"-"`
	Pulled bool `json:"-"`
}

type CustomColumn struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CustomColumns []CustomColumn

var (
	_ sql.Scanner   = (*CustomColumns)(nil)
	_ driver.Valuer = (*CustomColumns)(nil)
)

func (c *CustomColumns) Scan(value any) error {
	if value == nil {
		*c = CustomColumns{}
		return nil
	}

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal CustomColumns value %+v", value)
	}

	if len(data) == 0 {
		*c = CustomColumns{}
		return nil
	}

	return json.Unmarshal(data, c)
}

func (c CustomColumns) Value() (driver.Value, error) {
	sort.Slice(c, func(i, j int) bool {
		if c[i].Name == c[j].Name {
			return c[i].Value < c[j].Value
		}
		return c[i].Name < c[j].Name
	})

	return json.Marshal(c)
}
