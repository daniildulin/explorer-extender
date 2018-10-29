package models

import "time"

type Validator struct {
	ID                uint    `gorm:"primary_key"`
	Name              *string `json:"name"               gorm:"type:varchar(255)"`
	AccumulatedReward string  `json:"accumulated_reward" gorm:"type:numeric(300)"`
	AbsentTimes       uint    `json:"absent_times"`
	Address           string  `json:"address"`
	TotalStake        string  `json:"total_stake"        gorm:"type:numeric(300)"`
	PublicKey         string  `json:"public_key"         gorm:"type:varchar(255);unique_index"`
	Commission        uint    `json:"commission"`
	CreatedAtBlock    uint    `json:"created_at_block"`
	Status            byte    `json:"status"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}
