package models

import "time"

type RawReport struct {
	ID            uint `gorm:"primaryKey"`
	ReceiptNumber string
	CorpCode      string
	BlobData      []byte
	BlobSize      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
