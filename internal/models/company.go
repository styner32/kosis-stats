package models

import "time"

type Company struct {
	ID               uint `gorm:"primaryKey"`
	CorpCode         string
	CorpName         string
	CorpEngName      string
	LastModifiedDate time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
