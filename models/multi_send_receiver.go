package models

import "time"

type MultiSendReceiver struct {
	ID            uint64 `gorm:"primary_key"`
	TransactionID uint64 `json:"transaction_id" gorm:"type:bigint"`
	Coin          string `json:"coin"`
	To            string `json:"to"`
	Value         string `json:"value" gorm:"type:numeric(300,0)"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
