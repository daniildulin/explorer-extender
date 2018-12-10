package minter_api

import (
	"time"
)

type Response struct {
	JSONrpc string  `json:"jsonrpc"`
	ID      string  `json:"id"`
	Code    uint    `json:"code"`
	Log     *string `json:"log"`
}

type StatusResponse struct {
	Response
	Result StatusResult `json:"result"`
}

type StatusResult struct {
	LatestBlockHeight string `json:"latest_block_height"`
	LatestBlockHash   []byte `json:"latest_block_hash"`
	LatestAppHash     []byte `json:"latest_app_hash"`
}

type BlockResponse struct {
	Response
	Result BlockResult `json:"result"`
}

type ValidatorsResponse struct {
	Response
	Result []Validator `json:"result"`
}

type CoinInfoResponse struct {
	Response
	Result CoinInfoResult `json:"result"`
}

type BalanceResponse struct {
	Response
	Result BalanceResult `json:"result"`
}

type EventsResponse struct {
	Response
	Result EventsResult `json:"result"`
}

type BlockResult struct {
	Hash         string        `json:"hash"`
	Height       string        `json:"height"`
	TxCount      string        `json:"num_txs"`
	TotalTx      string        `json:"total_txs"`
	BlockReward  string        `json:"block_reward"`
	Size         string        `json:"size"`
	Time         time.Time     `json:"time"`
	Transactions []transaction `json:"transactions"`
	Validators   []Validator   `json:"validators"`
}

type EventsResult struct {
	Events *[]event `json:"events"`
}

type transaction struct {
	Hash        string             `json:"hash"`
	From        string             `json:"from"`
	Type        uint8              `json:"type"`
	Nonce       string             `json:"nonce"`
	GasPrice    string             `json:"gas_price"`
	GasCoin     string             `json:"gas_coin"`
	GasUsed     string             `json:"gas_used"`
	Gas         string             `json:"gas"`
	Payload     string             `json:"payload"`
	ServiceData string             `json:"service_data"`
	Data        transactionData    `json:"data"`
	Tags        *map[string]string `json:"tags"`
	Log         *string            `json:"log"`
}

type transactionData struct {
	Coin                 *string `json:"coin"`
	To                   *string `json:"to"`
	Value                *string `json:"value"`
	CoinToSell           *string `json:"coin_to_sell"`
	CoinToBuy            *string `json:"coin_to_buy"`
	ValueToSell          *string `json:"value_to_sell"`
	ValueToBuy           *string `json:"value_to_buy"`
	Name                 *string `json:"name"`
	Symbol               *string `json:"symbol"`
	InitialAmount        *string `json:"initial_amount"`
	InitialReserve       *string `json:"initial_reserve"`
	ConstantReserveRatio *string `json:"constant_reserve_ratio"`
	Address              *string `json:"address"`
	PubKey               *string `json:"pub_key"`
	Commission           *string `json:"commission"`
	Stake                *string `json:"stake"`
	Proof                *string `json:"proof"`
	RawCheck             *string `json:"raw_check"`
	ToCoinSymbol         *string `json:"to_coin_symbol"`
	FromCoinSymbol       *string `json:"from_coin_symbol"`
	Threshold            *string `json:"threshold"`
}

type event struct {
	Type  string     `json:"type"`
	Value eventValue `json:"value"`
}

type eventValue struct {
	Role            string `json:"role"`
	Coin            string `json:"coin"`
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	ValidatorPubKey string `json:"validator_pub_key"`
}

type Validator struct {
	PubKey string `json:"pubkey"`
}

type candidate struct {
	CandidateAddress string `json:"candidate_address"`
	TotalStake       string `json:"total_stake"`
	PubKey           string `json:"pub_key"`
	Commission       uint   `json:"commission"`
	CreatedAtBlock   uint   `json:"created_at_block"`
	Status           byte   `json:"status"`
	Stakes           []stake
}

type stake struct {
	Owner    string `json:"owner"`
	Coin     string `json:"coin"`
	Value    string `json:"value"`
	BipValue string `json:"bip_value"`
}

type CoinInfoResult struct {
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	Volume         string `json:"volume"`
	Crr            uint   `json:"crr"`
	ReserveBalance string `json:"reserve_balance"`
}

type BalanceResult struct {
	Balance map[string]string `json:"balance"`
}
