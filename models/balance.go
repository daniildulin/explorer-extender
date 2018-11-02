package models

import "time"

type Balance struct {
	ID        uint   `gorm:"primary_key"`
	Address   string `json:"address" gorm:"type:varchar(255)"`
	Coin      string `json:"coin"    gorm:"type:varchar(255)"`
	Amount    string `json:"amount"  gorm:"type:numeric(300)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
