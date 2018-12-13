package models

import "time"

type Validator struct {
	ID                uint64  `gorm:"primary_key"`
	Name              *string `json:"name"               gorm:"type:varchar(255)"`
	AccumulatedReward string  `json:"accumulated_reward" gorm:"type:numeric(300)"`
	AbsentTimes       uint64  `json:"absent_times"`
	Address           string  `json:"address"`
	TotalStake        string  `json:"total_stake"        gorm:"type:numeric(300)"`
	PublicKey         string  `json:"public_key"         gorm:"type:varchar(255);unique_index"`
	Commission        uint8   `json:"commission"`
	CreatedAtBlock    uint64  `json:"created_at_block"`
	Status            byte    `json:"status"`
	Stakes            []Stake
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}
