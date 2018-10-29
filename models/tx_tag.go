package models

import "time"

type TxTag struct {
	ID            uint `gorm:"primary_key"`
	TransactionID uint
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
