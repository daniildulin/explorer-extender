package models

import (
	"time"
)

type Stake struct {
	ID          uint64     `json:"-"            gorm:"primary_key"`
	ValidatorID uint64     `json:"validator_id" gorm:"primary_key"`
	Owner       string     `json:"owner"        gorm:"type:varchar(100)"`
	Coin        string     `json:"coin"         gorm:"type:varchar(100)"`
	Value       string     `json:"value"        gorm:"type:numeric(300,0)"`
	BipValue    string     `json:"bip_value"    gorm:"type:numeric(300,0)"`
	CreatedAt   time.Time  `json:"timestamp"`
	UpdatedAt   time.Time  `json:"-"`
	DeletedAt   *time.Time `json:"-"`
}
