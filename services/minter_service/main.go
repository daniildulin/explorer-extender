package minter_service

import (
	"encoding/base64"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/daniildulin/minter-node-api"
	"github.com/daniildulin/minter-node-api/responses"
	"github.com/grokify/html-strip-tags-go"
	"github.com/jinzhu/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

type events struct {
	Rewards []models.Reward
	Slashes []models.Slash
}

type MinterService struct {
	db     *gorm.DB
	config env.Config
	api    *minter_node_api.MinterNodeApi
	bs     *MinterBroadcastService
}

func New(config env.Config, db *gorm.DB, wsClient *MinterBroadcastService) *MinterService {

	apiLink := `http`
	if config.GetBool(`minterApi.isSecure`) {
		apiLink += `s://` + config.GetString(`minterApi.link`) + `:` + config.GetString(`minterApi.port`)
	} else {
		apiLink += `://` + config.GetString(`minterApi.link`) + `:` + config.GetString(`minterApi.port`)
	}

	return &MinterService{
		db:     db,
		config: config,
		api:    minter_node_api.New(apiLink),
		bs:     wsClient,
	}
}

func (ms *MinterService) Run() {

	currentDBBlock := ms.getLastBlockFromDB()
	lastApiBlock := ms.getLastBlockFromNode()

	if currentDBBlock >= 1 {
		ms.deleteBlockData(currentDBBlock)
	} else {
		currentDBBlock = 1
	}

	log.Printf("Connect to %s", ms.api.GetLink())
	log.Printf("Start from block %d", currentDBBlock)

	for {
		if currentDBBlock <= lastApiBlock {
			start := time.Now()
			ms.storeDataToDb(currentDBBlock)
			elapsed := time.Since(start)
			currentDBBlock++
			if ms.config.GetBool(`debug`) {
				log.Printf("Time of processing %s for block %s", elapsed, fmt.Sprint(currentDBBlock))
			}
		} else {
			lastApiBlock = ms.getLastBlockFromNode()
		}
	}
}

//func (ms *MinterService) GetActiveNodesCount() int {
//	return ms.api.GetActiveNodesCount()
//}
//func (ms *MinterService) UpdateApiNodesList() {
//	ms.api.GetActualNodes()
//}

func (ms *MinterService) getLastBlockFromNode() uint64 {
	status, err := ms.api.GetStatus()
	helpers.CheckErr(err)
	lastApiBlock, err := strconv.ParseUint(status.Result.LatestBlockHeight, 10, 64)
	helpers.CheckErr(err)
	return lastApiBlock
}

func (ms *MinterService) getLastBlockFromDB() uint64 {
	var b models.Block
	ms.db.Last(&b)
	return b.Height
}

func (ms *MinterService) deleteBlockData(blockHeight uint64) {
	if blockHeight > 0 {
		ms.db.Exec(`DELETE FROM blocks WHERE id=?`, blockHeight)
	}
}

func (ms *MinterService) storeDataToDb(blockHeight uint64) error {
	blockData, err := ms.api.GetBlock(blockHeight)
	if err != nil {
		return err
	}
	ms.storeBlockToDB(blockData)
	ms.storeBlockEvents(blockHeight)
	if ms.config.GetBool(`debug`) {
		log.Printf("Block: %s; Txs: %s; Hash: %s", blockData.Result.Height, blockData.Result.TxCount, blockData.Result.Hash)
	}
	return nil
}

func (ms *MinterService) storeBlockToDB(br *responses.BlockResponse) {
	blockData := br.Result
	height, err := strconv.Atoi(blockData.Height)
	helpers.CheckErr(err)
	if height <= 0 {
		return
	}

	txCount, err := strconv.Atoi(blockData.TxCount)
	helpers.CheckErr(err)
	size, err := strconv.Atoi(blockData.Size)
	helpers.CheckErr(err)

	blockModel := models.Block{
		ID:          uint64(height),
		Hash:        `Mh` + strings.ToLower(blockData.Hash),
		Height:      uint64(height),
		TxCount:     uint16(txCount),
		CreatedAt:   blockData.Time,
		Timestamp:   blockData.Time.UnixNano(),
		Size:        uint16(size),
		BlockTime:   ms.getBlockTime(uint64(height), blockData.Time),
		BlockReward: blockData.BlockReward,
	}

	if blockModel.TxCount > 0 {
		blockModel.Transactions = ms.getTransactionModelsFromApiData(br)
		for i, tx := range blockModel.Transactions {
			if i <= 19 {
				go ms.bs.Transaction(&tx)
			}
		}
	}

	blockModel.Validators = ms.getValidatorModels(br)

	ms.db.Create(&blockModel)
	go ms.updateValidatorsInfo(&blockModel)
	go ms.bs.Block(&blockModel)
}

func (ms *MinterService) storeBlockEvents(blockHeight uint64) {
	response, err := ms.api.GetBlockEvents(blockHeight)
	helpers.CheckErr(err)
	events := getEventsModelsFromApiData(response, blockHeight)
	for _, e := range events.Slashes {
		ms.db.Create(&e)
	}
	for _, e := range events.Rewards {
		ms.db.Create(&e)
	}
}

func (ms *MinterService) getBlockTime(currentBlockHeight uint64, blockTime time.Time) float64 {

	if currentBlockHeight == 1 {
		return 1
	}

	var b models.Block
	ms.db.Where("height = ?", currentBlockHeight-1).First(&b)

	result := blockTime.Sub(b.CreatedAt)
	if result < 0 {
		return 5
	}

	return result.Seconds()
}

func (ms *MinterService) getValidatorModels(response *responses.BlockResponse) []models.Validator {
	var result []models.Validator
	blockData := response.Result
	for _, v := range blockData.Validators {
		var vld models.Validator
		ms.db.Where("public_key = ?", v.PubKey).First(&vld)
		if vld.ID == 0 {
			result = append(result, models.Validator{
				Name:              nil,
				AccumulatedReward: `0`,
				AbsentTimes:       0,
				Address:           ``,
				TotalStake:        `0`,
				PublicKey:         v.PubKey,
				Commission:        0,
				CreatedAtBlock:    0,
				Status:            2,
			})
		} else if vld.ID != 0 {
			result = append(result, vld)
		}

	}
	return result
}

func (ms *MinterService) getTransactionModelsFromApiData(response *responses.BlockResponse) []models.Transaction {

	blockData := response.Result

	txCount, err := strconv.Atoi(blockData.TxCount)
	helpers.CheckErr(err)

	var result = make([]models.Transaction, txCount)

	for i, tx := range blockData.Transactions {

		status := tx.Log == nil
		payload, _ := base64.StdEncoding.DecodeString(tx.Payload)

		nonce, err := strconv.ParseUint(tx.Nonce, 10, 64)
		helpers.CheckErr(err)
		gasPrice, err := strconv.ParseUint(tx.GasPrice, 10, 64)
		helpers.CheckErr(err)
		gas, err := strconv.ParseUint(tx.Gas, 10, 64)
		helpers.CheckErr(err)

		transaction := models.Transaction{
			Hash:                 strings.Title(tx.Hash),
			From:                 strings.Title(tx.From),
			Type:                 tx.Type,
			Nonce:                nonce,
			GasPrice:             gasPrice,
			GasCoin:              &tx.GasCoin,
			Fee:                  gas,
			Payload:              strip.StripTags(string(payload[:])),
			ServiceData:          strip.StripTags(tx.ServiceData),
			CreatedAt:            blockData.Time,
			To:                   tx.Data.To,
			Address:              tx.Data.Address,
			Name:                 stripHtmlTags(tx.Data.Name),
			CoinToSell:           tx.Data.CoinToSell,
			CoinToBuy:            tx.Data.CoinToBuy,
			Stake:                tx.Data.Stake,
			Value:                tx.Data.Value,
			Commission:           nil,
			InitialAmount:        tx.Data.InitialAmount,
			InitialReserve:       tx.Data.InitialReserve,
			ConstantReserveRatio: nil,
			RawCheck:             tx.Data.RawCheck,
			Proof:                tx.Data.Proof,
			Coin:                 stripHtmlTags(tx.Data.Coin),
			PubKey:               tx.Data.PubKey,
			Status:               status,
			Threshold:            nil,
			Log:                  tx.Log,
		}

		if tx.Data.Commission != nil {
			commission, err := strconv.ParseUint(*tx.Data.Commission, 10, 64)
			helpers.CheckErr(err)
			transaction.Commission = &commission
		}

		if tx.Data.ConstantReserveRatio != nil {
			crr, err := strconv.ParseUint(*tx.Data.ConstantReserveRatio, 10, 64)
			helpers.CheckErr(err)
			transaction.ConstantReserveRatio = &crr
		}

		if tx.Data.Threshold != nil {
			threshold, err := strconv.ParseUint(*tx.Data.Threshold, 10, 64)
			helpers.CheckErr(err)
			transaction.Threshold = &threshold
		}

		var tags []models.TxTag
		if tx.Tags != nil {
			for k, tg := range *tx.Tags {
				tags = append(tags, models.TxTag{
					Key:   k,
					Value: tg,
				})
			}
			transaction.Tags = tags
		}

		if tx.Type == models.TX_TYPE_CREATE_COIN {
			transaction.Coin = tx.Data.Symbol
			ms.createFromTransaction(&transaction)
		}

		if tx.Type == models.TX_TYPE_SELL_COIN {
			transaction.ValueToSell = tx.Data.ValueToSell
			transaction.ValueToBuy = getValueFromTxTag(tags, `tx.return`)
		}
		if tx.Type == models.TX_TYPE_SELL_ALL_COIN {
			transaction.ValueToSell = getValueFromTxTag(tags, `tx.sell_amount`)
			transaction.ValueToBuy = getValueFromTxTag(tags, `tx.return`)
		}
		if tx.Type == models.TX_TYPE_BUY_COIN {
			transaction.ValueToSell = getValueFromTxTag(tags, `tx.return`)
			transaction.ValueToBuy = tx.Data.ValueToBuy
		}
		go ms.updateBalances(&transaction)
		go ms.updateCoins(&transaction)
		result[i] = transaction
	}

	return result
}

func (ms *MinterService) createFromTransaction(transaction *models.Transaction) {
	ms.db.Save(&models.Coin{
		Symbol:         *transaction.Coin,
		Name:           *transaction.Name,
		Volume:         *transaction.InitialAmount,
		ReserveBalance: *transaction.InitialReserve,
		Crr:            *transaction.ConstantReserveRatio,
		Creator:        transaction.From,
		CreatedAt:      transaction.CreatedAt,
	})
}

func (ms *MinterService) updateCoins(tx *models.Transaction) {
	for _, coin := range []*string{tx.GasCoin, tx.CoinToSell, tx.CoinToBuy} {
		ms.updateCoin(coin)
	}
}

func (ms *MinterService) updateCoin(coin *string) {
	if coin == nil || *coin == ms.config.GetString(`baseCoin`) {
		return
	}
	data, err := ms.api.GetCoinInfo(*coin)
	if err != nil {
		helpers.CheckErr(err)
	}

	if data.Error.Code == 404 {
		ms.db.Exec(`DELETE FROM coins WHERE symbol = ?`, coin)
		log.Printf(`Coin %s have been deleted`, *coin)
	} else {
		var c models.Coin
		ms.db.Where("symbol = ?", coin).First(&c)
		ms.db.Model(&c).Updates(map[string]interface{}{
			"volume":          data.Result.Volume,
			"reserve_balance": data.Result.ReserveBalance,
		})
		log.Printf(`Coin %s have been updated`, *coin)
	}
}

func (ms *MinterService) updateBalances(tx *models.Transaction) {
	ms.updateAddressBalance(tx.From)
	if tx.To != nil && *tx.To != tx.From {
		ms.updateAddressBalance(*tx.To)
	}
}

func (ms *MinterService) updateAddressBalance(address string) {
	var coinsList []string
	data, _ := ms.api.GetAddress(address)
	for coin, amount := range data.Result.Balance {
		coinsList = append(coinsList, coin)
		var balance models.Balance
		ms.db.Where("address = ? AND  coin = ?", address, coin).First(&balance)

		if balance.ID == 0 {
			balance.Address = address
			balance.Coin = coin
			balance.Amount = amount
			ms.db.Create(&balance)
		} else {
			ms.db.Model(&balance).Update("amount", amount)
		}
		go ms.bs.Balance(&balance)
	}
	ms.db.Exec(`DELETE FROM balances WHERE address = ? AND coin NOT IN (?) `, address, coinsList)
}

func (ms *MinterService) updateValidatorsInfo(blockModel *models.Block) {
	for _, candidate := range blockModel.Validators {
		ms.updateValidatorInfo(candidate.PublicKey)
	}
}

func (ms *MinterService) updateValidatorInfo(pubKey string) {
	var validator models.Validator

	response, err := ms.api.GetCandidate(pubKey)
	helpers.CheckErr(err)
	ms.db.Where("public_key = ?", pubKey).First(&validator)

	if validator.ID != 0 {
		commission, err := strconv.ParseUint(response.Result.Commission, 10, 8)
		helpers.CheckErr(err)
		createdAtBlock, err := strconv.ParseUint(response.Result.CreatedAtBlock, 10, 64)
		helpers.CheckErr(err)
		validator.Address = response.Result.CandidateAddress
		validator.TotalStake = response.Result.TotalStake
		validator.Commission = uint8(commission)
		validator.CreatedAtBlock = createdAtBlock
		validator.Status = response.Result.Status
		ms.db.Save(&validator)
	}
}

func stripHtmlTags(str *string) *string {
	var s string
	if str != nil {
		s = *str
		s = strip.StripTags(s)
	}
	return &s
}

func getValueFromTxTag(tags []models.TxTag, tagName string) *string {
	for _, tag := range tags {
		if tag.Key == tagName {
			return &tag.Value
		}
	}
	return nil
}

func getEventsModelsFromApiData(response *responses.EventsResponse, blockHeight uint64) events {
	var rewards []models.Reward
	var slashes []models.Slash
	eventData := response.Result

	if eventData.Events != nil {
		for _, event := range *eventData.Events {
			if event.Type == `minter/RewardEvent` {
				rewards = append(rewards, models.Reward{
					BlockID:     uint(blockHeight),
					Role:        event.Value.Role,
					Amount:      event.Value.Amount,
					Address:     event.Value.Address,
					ValidatorPk: event.Value.ValidatorPubKey,
				})
			} else if event.Type == `minter/SlashEvent` {
				slashes = append(slashes, models.Slash{
					BlockID:     uint(blockHeight),
					Coin:        event.Value.Coin,
					Amount:      event.Value.Amount,
					Address:     event.Value.Address,
					ValidatorPk: event.Value.ValidatorPubKey,
				})
			}
		}
	}

	return events{
		Rewards: rewards,
		Slashes: slashes,
	}
}
