package models

import (
	"time"
)

const TX_TYPE_SEND = 1
const TX_TYPE_SELL_COIN = 2
const TX_TYPE_SELL_ALL_COIN = 3
const TX_TYPE_BUY_COIN = 4
const TX_TYPE_CREATE_COIN = 5
const TX_TYPE_DECLARE_CANDIDACY = 6
const TX_TYPE_DELEGATE = 7
const TX_TYPE_UNBOUND = 8
const TX_TYPE_REDEEM_CHECK = 9
const TX_TYPE_SET_CANDIDATE_ONLINE = 10
const TX_TYPE_SET_CANDIDATE_OFFLINE = 11
const TX_TYPE_MULTI_SIG = 12

type Transaction struct {
	ID                   uint    `gorm:"primary_key"`
	BlockID              uint    `json:"block_id"               gorm:"type:bigint"`
	Type                 uint    `json:"type"`
	From                 string  `json:"from"                   gorm:"type:varchar(100)"`
	To                   *string `json:"to"                     gorm:"type:varchar(100)"`
	Hash                 string  `json:"hash"                   gorm:"type:varchar(100)"`
	PubKey               *string `json:"pub_key"                gorm:"type:varchar(255)"`
	Value                *string `json:"value"                  gorm:"type:numeric(300,0)"`
	ValueToSell          *string `json:"value_to_sell"          gorm:"type:numeric(300,0)"`
	ValueToBuy           *string `json:"value_to_buy"           gorm:"type:numeric(300,0)"`
	Fee                  uint    `json:"fee"                    gorm:"type:numeric(300,0)"`
	Stake                *string `json:"stake"                  gorm:"type:numeric(300,0)"`
	Commission           *uint   `json:"Commission"             gorm:"type:numeric(300,0)"`
	InitialAmount        *string `json:"initial_amount"         gorm:"type:numeric(300,0)"`
	InitialReserve       *string `json:"initial_reserve"        gorm:"type:numeric(50,0)"`
	ConstantReserveRatio *uint   `json:"constant_reserve_ratio" gorm:"type:numeric(300,0)"`
	GasWanted            *string `json:"gas_wanted"             gorm:"type:numeric(300,0)"`
	GasUsed              *string `json:"gas_used"               gorm:"type:numeric(300,0)"`
	GasPrice             uint    `json:"gas_price"              gorm:"type:numeric(300,0)"`
	GasCoin              *string `json:"gas_coin"               gorm:"type:varchar(20)"`
	Coin                 *string `json:"coin"                   gorm:"type:varchar(255)"`
	Nonce                uint    `json:"nonce"`
	Threshold            *uint   `json:"threshold"`
	Payload              string  `json:"payload"                gorm:"type:text"`
	ServiceData          string  `json:"service_data"           gorm:"type:text"`
	Address              *string `json:"Address"                gorm:"type:varchar(255)"`
	CoinToSell           *string `json:"coin_to_sell"           gorm:"type:varchar(25)"`
	CoinToBuy            *string `json:"coin_to_buy"            gorm:"type:varchar(25)"`
	RawCheck             *string `json:"raw_check"              gorm:"type:text"`
	Proof                *string `json:"proof"`
	Name                 *string `json:"name"`
	Log                  *string `json:"log"`
	Status               bool    `json:"status"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
	Tags                 []TxTag
}
