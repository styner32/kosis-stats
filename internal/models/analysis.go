package models

import (
	"encoding/json"
	"time"
)

type Analysis struct {
	ID          uint `gorm:"primaryKey"`
	RawReportID uint
	Analysis    json.RawMessage `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
