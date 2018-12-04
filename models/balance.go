package models

import "time"

type Balance struct {
	ID        uint       `json:"-" gorm:"primary_key"`
	Address   string     `json:"address" gorm:"type:varchar(255)"`
	Coin      string     `json:"coin"    gorm:"type:varchar(255)"`
	Amount    string     `json:"amount"  gorm:"type:numeric(300)"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}
