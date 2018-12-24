package models

import (
	"github.com/daniildulin/explorer-extender/helpers"
	"math/big"
	"time"
)

type Block struct {
	ID           uint64        `json:"-"         gorm:"primary_key"`
	Height       uint64        `json:"height"    gorm:"type:bigint;unique_index"`
	Timestamp    int64         `json:"-"         gorm:"type:bigint"`
	TxCount      uint16        `json:"txCount"`
	TxTotal      uint64        `json:"txTotal"`
	Size         uint16        `json:"size"`
	BlockTime    float64       `json:"blockTime" gorm:"type:numeric(20, 10)"`
	Hash         string        `json:"hash"      gorm:"type:varchar(255)"`
	BlockReward  string        `json:"reward"    gorm:"type:numeric(300, 0)"`
	Validators   []Validator   `json:"-"         gorm:"many2many:block_validator;"`
	Transactions []Transaction `json:"-"`
	Rewards      []Reward      `json:"-"`
	Slashes      []Slash       `json:"-"`
	CreatedAt    time.Time     `json:"timestamp"`
	UpdatedAt    time.Time     `json:"-"`
	DeletedAt    *time.Time    `json:"-"`
}

func NewBlock(ID uint64, height uint64, timestamp int64, txCount uint16, totalTx uint64, size uint16, blockTime float64,
	hash string, blockReward string, validators []Validator, transactions []Transaction, rewards []Reward,
	slashes []Slash, createdAt time.Time, updatedAt time.Time, deletedAt *time.Time) *Block {
	return &Block{
		ID:           ID,
		Height:       height,
		Timestamp:    timestamp,
		TxCount:      txCount,
		TxTotal:      totalTx,
		Size:         size,
		BlockTime:    blockTime,
		Hash:         hash,
		BlockReward:  blockReward,
		Validators:   validators,
		Transactions: transactions,
		Rewards:      rewards,
		Slashes:      slashes,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt:    deletedAt,
	}
}

type BlockResponse struct {
	Hash              string    `json:"hash"`
	Height            uint64    `json:"height"`
	Reward            string    `json:"reward"`
	Size              uint32    `json:"size"`
	TxCount           uint32    `json:"txCount"`
	TxTotal           uint64    `json:"txTotal"`
	Timestamp         time.Time `json:"timestamp"`
	LatestBlockHeight uint32    `json:"latestBlockHeight"`
	BlockTime         float64   `json:"blockTime"`
}

func NewBlockResponse(hash string, height uint64, reward string, size uint32, txCount uint32, txTotal uint64, timestamp time.Time,
	latestBlockHeight uint32, blockTime float64) *BlockResponse {
	return &BlockResponse{
		Hash:              hash,
		Height:            height,
		Reward:            reward,
		Size:              size,
		TxCount:           txCount,
		TxTotal:           txTotal,
		Timestamp:         timestamp,
		LatestBlockHeight: latestBlockHeight,
		BlockTime:         blockTime,
	}
}

func (b *Block) GetReward() *big.Float {
	return helpers.PipValueToCoin(b.BlockReward)
}

func (b *Block) GetResponse() *BlockResponse {
	return NewBlockResponse(b.Hash, uint64(b.Height), b.GetReward().String(), uint32(b.Size), uint32(b.TxCount),
		b.TxTotal, b.CreatedAt, 0, b.BlockTime)
}
