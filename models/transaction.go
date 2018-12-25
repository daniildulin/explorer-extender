package models

import (
	"fmt"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/jinzhu/gorm/dialects/postgres"
	"math/big"
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
const TX_TYPE_MULTI_SEND = 13

var txType = [14]string{
	`-`,
	`send`,
	`sellCoin`,
	`sellAllCoin`,
	`buyCoin`,
	`createCoin`,
	`declareCandidacy`,
	`delegate`,
	`unbond`,
	`redeemCheckData`,
	`setCandidateOnData`,
	`setCandidateOffData`,
	`multiSig`,
	`multiSend`,
}

type Transaction struct {
	ID                   uint64          `gorm:"primary_key"`
	BlockID              uint64          `json:"block_id"               gorm:"type:bigint"`
	Type                 uint8           `json:"type"`
	From                 string          `json:"from"                   gorm:"type:varchar(100)"`
	To                   *string         `json:"to"                     gorm:"type:varchar(100)"`
	Hash                 string          `json:"hash"                   gorm:"type:varchar(100)"`
	PubKey               *string         `json:"pub_key"                gorm:"type:varchar(255)"`
	Value                *string         `json:"value"                  gorm:"type:numeric(300,0)"`
	ValueToSell          *string         `json:"value_to_sell"          gorm:"type:numeric(300,0)"`
	ValueToBuy           *string         `json:"value_to_buy"           gorm:"type:numeric(300,0)"`
	Fee                  uint64          `json:"fee"                    gorm:"type:numeric(300,0)"`
	Stake                *string         `json:"stake"                  gorm:"type:numeric(300,0)"`
	Commission           *uint64         `json:"Commission"             gorm:"type:numeric(300,0)"`
	InitialAmount        *string         `json:"initial_amount"         gorm:"type:numeric(300,0)"`
	InitialReserve       *string         `json:"initial_reserve"        gorm:"type:numeric(50,0)"`
	ConstantReserveRatio *uint64         `json:"constant_reserve_ratio" gorm:"type:numeric(300,0)"`
	GasWanted            *string         `json:"gas_wanted"             gorm:"type:numeric(300,0)"`
	GasUsed              *string         `json:"gas_used"               gorm:"type:numeric(300,0)"`
	GasPrice             uint64          `json:"gas_price"              gorm:"type:numeric(300,0)"`
	GasCoin              *string         `json:"gas_coin"               gorm:"type:varchar(20)"`
	Coin                 *string         `json:"coin"                   gorm:"type:varchar(255)"`
	Nonce                uint64          `json:"nonce"`
	Threshold            *uint64         `json:"threshold"`
	Payload              string          `json:"payload"                gorm:"type:text"`
	ServiceData          string          `json:"service_data"           gorm:"type:text"`
	Address              *string         `json:"Address"                gorm:"type:varchar(255)"`
	CoinToSell           *string         `json:"coin_to_sell"           gorm:"type:varchar(25)"`
	CoinToBuy            *string         `json:"coin_to_buy"            gorm:"type:varchar(25)"`
	RawCheck             *string         `json:"raw_check"              gorm:"type:text"`
	Check                *postgres.Jsonb `json:"check"`
	Proof                *string         `json:"proof"`
	Name                 *string         `json:"name"`
	Log                  *string         `json:"log"`
	Status               bool            `json:"status"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
	Tags                 []TxTag
	MultiSendReceivers   []MultiSendReceiver
}

type TransactionResponse struct {
	Txn       uint64            `json:"txn"`
	Hash      string            `json:"hash"`
	Nonce     uint64            `json:"nonce"`
	Block     uint64            `json:"block"`
	Timestamp time.Time         `json:"timestamp"`
	Fee       string            `json:"fee"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Payload   string            `json:"payload"`
	From      string            `json:"from"`
	Data      map[string]string `json:"data"`
}

func (tx Transaction) GetFee() *big.Float {
	bip := big.NewFloat(0.000000000000000001)
	value := big.NewFloat(float64(tx.Fee))
	return value.Mul(value, bip)
}

func (tx Transaction) GetTypeString() string {

	if int(tx.Type) > len(txType)+1 {
		return `Unknown transaction type`
	}

	return txType[tx.Type]
}

func (tx Transaction) GetStatusString() string {
	if tx.Status {
		return `success`
	}
	return `fail`
}

func (tx Transaction) GetResponse() *TransactionResponse {

	var data = make(map[string]string)

	if tx.Type == TX_TYPE_SEND {
		if tx.To != nil {
			data[`to`] = *tx.To
		}
		if tx.Coin != nil {
			data[`coin`] = *tx.Coin
		}
		if tx.Value != nil {
			data[`amount`] = helpers.PipValueToCoin(*tx.Value).String()
		}
	}

	if tx.Type == TX_TYPE_SELL_COIN || tx.Type == TX_TYPE_SELL_ALL_COIN || tx.Type == TX_TYPE_BUY_COIN {
		if tx.CoinToSell != nil {
			data[`coin_to_sell`] = *tx.CoinToSell
		}
		if tx.CoinToBuy != nil {
			data[`coin_to_buy`] = *tx.CoinToBuy
		}
		if tx.ValueToBuy != nil {
			data[`value_to_buy`] = helpers.PipValueToCoin(*tx.ValueToBuy).String()
		}
		if tx.ValueToSell != nil {
			data[`value_to_sell`] = helpers.PipValueToCoin(*tx.ValueToSell).String()
		}
	}

	if tx.Type == TX_TYPE_CREATE_COIN {
		if tx.Name != nil {
			data[`name`] = *tx.Name
		}
		if tx.Coin != nil {
			data[`symbol`] = *tx.Coin
		}
		if tx.InitialAmount != nil {
			data[`initial_amount`] = helpers.PipValueToCoin(*tx.InitialAmount).String()
		}
		if tx.InitialReserve != nil {
			data[`initial_reserve`] = helpers.PipValueToCoin(*tx.InitialReserve).String()
		}
		if tx.ConstantReserveRatio != nil {
			data[`constant_reserve_ratio`] = fmt.Sprint(*tx.ConstantReserveRatio)
		}
	}

	if tx.Type == TX_TYPE_DECLARE_CANDIDACY {
		if tx.Address != nil {
			data[`address`] = *tx.Address
		}
		if tx.PubKey != nil {
			data[`pub_key`] = *tx.PubKey
		}
		if tx.Commission != nil {
			data[`commission`] = fmt.Sprint(*tx.Commission)
		}
		if tx.Coin != nil {
			data[`coin`] = *tx.Coin
		}
		if tx.Stake != nil {
			data[`stake`] = helpers.PipValueToCoin(*tx.Stake).String()
		}
	}

	if tx.Type == TX_TYPE_DELEGATE || tx.Type == TX_TYPE_UNBOUND {
		//$value = $this->value ?? $this->stake ?? 0; --- UNBOUND
		if tx.PubKey != nil {
			data[`pub_key`] = *tx.PubKey
		}
		if tx.Coin != nil {
			data[`coin`] = *tx.Coin
		}
		if tx.Stake != nil {
			data[`stake`] = helpers.PipValueToCoin(*tx.Stake).String()
		}
	}

	if tx.Type == TX_TYPE_REDEEM_CHECK {
		if tx.PubKey != nil {
			data[`raw_check`] = *tx.RawCheck
		}
		if tx.Proof != nil {
			data[`proof`] = *tx.Proof
		}
	}

	if tx.Type == TX_TYPE_SET_CANDIDATE_ONLINE || tx.Type == TX_TYPE_SET_CANDIDATE_OFFLINE {
		if tx.PubKey != nil {
			data[`pub_key`] = *tx.PubKey
		}
	}

	r := &TransactionResponse{
		tx.ID,
		tx.Hash,
		tx.Nonce,
		tx.BlockID,
		tx.CreatedAt,
		tx.GetFee().String(),
		tx.GetTypeString(),
		tx.GetStatusString(),
		tx.Payload,
		tx.From,
		data,
	}

	return r
}
