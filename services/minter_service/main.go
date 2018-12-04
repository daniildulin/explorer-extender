package minter_service

import (
	"encoding/base64"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/daniildulin/explorer-extender/services/minter_api"
	"github.com/grokify/html-strip-tags-go"
	"github.com/jinzhu/gorm"
	"log"
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
	api    *minter_api.MinterApi
	bs     *MinterBroadcastService
}

func New(config env.Config, db *gorm.DB, minterApi *minter_api.MinterApi, wsClient *MinterBroadcastService) *MinterService {
	return &MinterService{
		db:     db,
		config: config,
		api:    minterApi,
		bs:     wsClient,
	}
}

func (ms *MinterService) Run() {

	currentDBBlock := ms.getLastBlockFromDB()
	lastApiBlock, _ := ms.api.GetLastBlock()

	if currentDBBlock >= 1 {
		ms.deleteBlockData(currentDBBlock)
	} else {
		currentDBBlock = 1
	}

	log.Printf("Connect to %s", ms.config.GetString("minterApi.link"))
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
			lastApiBlock, _ = ms.api.GetLastBlock()
		}
		//time.Sleep(300 * time.Millisecond)
	}
}

func (ms *MinterService) GetActiveNodesCount() int {
	return ms.api.GetActiveNodesCount()
}

func (ms *MinterService) UpdateApiNodesList() {
	ms.api.GetActualNodes()
}

func (ms *MinterService) getLastBlockFromDB() uint {
	var b models.Block
	ms.db.Last(&b)
	return b.Height
}

func (ms *MinterService) deleteBlockData(blockHeight uint) {
	if blockHeight > 0 {
		ms.db.Exec(`DELETE FROM blocks WHERE id=?`, blockHeight)
	}
}

func (ms *MinterService) storeDataToDb(blockHeight uint) error {
	blockData, err := ms.api.GetBlock(blockHeight)
	if err != nil {
		return err
	}
	ms.storeBlockToDB(&blockData.Result)
	//ms.storeBlockValidators(blockHeight)
	if ms.config.GetBool(`debug`) {
		log.Printf("Block: %d; Txs: %d; Hash: %s", blockData.Result.Height, blockData.Result.TxCount, blockData.Result.Hash)
	}
	return nil
}

func (ms *MinterService) storeBlockToDB(blockData *minter_api.BlockResult) {

	if blockData.Height <= 0 {
		return
	}

	blockModel := models.Block{
		ID:          blockData.Height,
		Hash:        `Mh` + strings.ToLower(blockData.Hash),
		Height:      blockData.Height,
		TxCount:     blockData.TxCount,
		CreatedAt:   blockData.Time,
		Timestamp:   blockData.Time.UnixNano(),
		Size:        blockData.Size,
		BlockTime:   ms.getBlockTime(blockData.Height, blockData.Time),
		BlockReward: blockData.BlockReward,
	}

	if blockModel.TxCount > 0 {
		blockModel.Transactions = ms.getTransactionModelsFromApiData(blockData)
		for i, tx := range blockModel.Transactions {
			if i <= 19 {
				go ms.bs.Transaction(&tx)
			}
		}
	}

	if blockData.Events != nil {
		e := getEventsModelsFromApiData(blockData)
		blockModel.Rewards = e.Rewards
		blockModel.Slashes = e.Slashes
	}

	ms.db.Create(&blockModel)
	go ms.bs.Block(&blockModel)
}

func (ms *MinterService) storeBlockValidators(blockHeight uint) {

	response, err := ms.api.GetBlockValidators(blockHeight)
	helpers.CheckErr(err)

	validators := ms.getValidatorModels(response.Result)

	var block models.Block
	ms.db.First(&block, blockHeight)

	// begin a transaction
	for _, v := range validators {
		if v.ID != 0 {
			ms.db.Save(&v)
			ms.db.Exec(`INSERT INTO block_validator (block_id, validator_id) VALUES (?, ?)`, blockHeight, v.ID)
		} else {
			ms.db.Create(&v)
			ms.db.Exec(`INSERT INTO block_validator (block_id, validator_id) VALUES (?, ?)`, blockHeight, v.ID)
		}
	}
}

func (ms *MinterService) getBlockTime(currentBlockHeight uint, blockTime time.Time) float64 {

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

func (ms *MinterService) getValidatorModels(validatorsData []minter_api.Validator) []models.Validator {

	var result []models.Validator

	for _, v := range validatorsData {

		var vld models.Validator

		ms.db.Where("public_key = ?", v.Candidate.PubKey).First(&vld)

		if vld.ID == 0 {
			result = append(result, models.Validator{
				Name:              nil,
				AccumulatedReward: v.AccumulatedReward,
				AbsentTimes:       v.AbsentTimes,
				Address:           v.Candidate.CandidateAddress,
				TotalStake:        v.Candidate.TotalStake,
				PublicKey:         v.Candidate.PubKey,
				Commission:        v.Candidate.Commission,
				CreatedAtBlock:    v.Candidate.CreatedAtBlock,
				Status:            v.Candidate.Status,
			})
		} else if vld.ID != 0 {
			vld.TotalStake = v.Candidate.TotalStake
			vld.AccumulatedReward = v.AccumulatedReward
			vld.AbsentTimes = v.AbsentTimes
			vld.Status = v.Candidate.Status
			result = append(result, vld)
		}

	}
	return result
}

func (ms *MinterService) getTransactionModelsFromApiData(blockData *minter_api.BlockResult) []models.Transaction {

	var result = make([]models.Transaction, blockData.TxCount)

	for i, tx := range blockData.Transactions {

		status := tx.Log == nil
		payload, _ := base64.StdEncoding.DecodeString(tx.Payload)

		transaction := models.Transaction{
			Hash:                 strings.Title(tx.Hash),
			From:                 strings.Title(tx.From),
			Type:                 tx.Type,
			Nonce:                tx.Nonce,
			GasPrice:             tx.GasPrice,
			GasCoin:              &tx.GasCoin,
			Fee:                  tx.Gas,
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
			Commission:           tx.Data.Commission,
			InitialAmount:        tx.Data.InitialAmount,
			InitialReserve:       tx.Data.InitialReserve,
			ConstantReserveRatio: tx.Data.ConstantReserveRatio,
			RawCheck:             tx.Data.RawCheck,
			Proof:                tx.Data.Proof,
			Coin:                 stripHtmlTags(tx.Data.Coin),
			PubKey:               tx.Data.PubKey,
			Status:               status,
			Threshold:            tx.Data.Threshold,
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
		log.Println(err.Error())
	}

	if data.Code == 404 {
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
	data, _ := ms.api.GetAddressBalance(address)
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
			//ms.db.Exec(`UPDATE balances SET amount = ? WHERE id = ? `, amount, balance.ID)
		}
		go ms.bs.Balance(&balance)
	}
	ms.db.Exec(`DELETE FROM balances WHERE address = ? AND coin NOT IN (?) `, address, coinsList)
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

func getEventsModelsFromApiData(blockData *minter_api.BlockResult) events {
	var rewards []models.Reward
	var slashes []models.Slash
	for _, event := range blockData.Events {
		if event.Type == `minter/RewardEvent` {
			rewards = append(rewards, models.Reward{
				Role:        event.Value.Role,
				Amount:      event.Value.Amount,
				Address:     event.Value.Address,
				ValidatorPk: event.Value.ValidatorPubKey,
			})
		} else if event.Type == `minter/SlashEvent` {
			slashes = append(slashes, models.Slash{
				Coin:        event.Value.Coin,
				Amount:      event.Value.Amount,
				Address:     event.Value.Address,
				ValidatorPk: event.Value.ValidatorPubKey,
			})
		}
	}
	return events{
		Rewards: rewards,
		Slashes: slashes,
	}
}
