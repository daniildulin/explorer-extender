package models

import (
	"github.com/daniildulin/explorer-extender/helpers"
	"time"
)

type Balance struct {
	ID        uint       `json:"-" gorm:"primary_key"`
	Address   string     `json:"address" gorm:"type:varchar(255)"`
	Coin      string     `json:"coin"    gorm:"type:varchar(255)"`
	Amount    string     `json:"amount"  gorm:"type:numeric(300)"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}

type BalanceResponse struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Amount  string `json:"amount"`
}

func (b *Balance) GetResponse() *BalanceResponse {
	return &BalanceResponse{
		Address: b.Address,
		Coin:    b.Coin,
		Amount:  helpers.PipValueToCoin(b.Amount).String(),
	}
}
