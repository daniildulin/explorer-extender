package node

import (
	"time"
)

type StatusResponse struct {
	Code   uint         `json:"code"`
	Result StatusResult `json:"result"`
}

type StatusResult struct {
	LatestBlockHeight string `json:"latest_block_height"`
	LatestBlockHash   []byte `json:"latest_block_hash"`
	LatestAppHash     []byte `json:"latest_app_hash"`
}

type blockResponse struct {
	Code   uint        `json:"code"`
	Result blockResult `json:"result"`
}

type blockResult struct {
	Hash        string    `json:"hash"`
	Height      uint      `json:"height"`
	TxCount     uint      `json:"num_txs"`
	TotalTx     uint      `json:"total_txs"`
	BlockReward string    `json:"block_reward"`
	Size        uint      `json:"size"`
	Time        time.Time `json:"time"`
}
