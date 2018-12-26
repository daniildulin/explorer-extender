package models

import (
	"time"
)

type Reward struct {
	ID          uint   `gorm:"primary_key"`
	BlockID     uint   `json:"block_id"     gorm:"type:bigint"`
	Role        string `json:"role"         gorm:"type:varchar(255)"`
	Amount      string `json:"amount"       gorm:"type:numeric(300)"`
	Address     string `json:"address"      gorm:"type:varchar(255)"`
	ValidatorPk string `json:"validator_pk" gorm:"type:varchar(255)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func (r Reward) GetAddress() string {
	return r.Address
}
