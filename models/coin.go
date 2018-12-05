package models

import "time"

type Coin struct {
	ID             uint64 `gorm:"primary_key"`
	Symbol         string `json:"symbol"          gorm:"type:varchar(255)"`
	Name           string `json:"name"            gorm:"type:varchar(255)"`
	Crr            uint64 `json:"crr"             gorm:"type:int"`
	Volume         string `json:"volume"          gorm:"type:numeric(300)"`
	ReserveBalance string `json:"reserve_balance" gorm:"type:numeric(300)"`
	Creator        string `json:"creator"         gorm:"type:varchar(255)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
