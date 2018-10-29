package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/grokify/html-strip-tags-go"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 1 * time.Second}

type events struct {
	Rewards []models.Reward
	Slashes []models.Slash
}

func Run(config env.Config, db *gorm.DB) {

	currentDBBlock := getLastBlockFromDB(db) + 1
	lastApiBlock := getLastBlockFromMinterAPI(config)
	log.Printf("Connect to %s", config.GetString("minterApi"))
	log.Printf("Start from block %d", currentDBBlock)

	for {
		if currentDBBlock <= lastApiBlock {
			start := time.Now()
			storeDataToDb(config, db, currentDBBlock)
			elapsed := time.Since(start)
			currentDBBlock++

			if config.GetBool(`debug`) {
				log.Printf("Time of processing %s for block %s", elapsed, fmt.Sprint(currentDBBlock))
			}

		} else {
			lastApiBlock = getLastBlockFromMinterAPI(config)
		}
	}
}

func getApiLink(config env.Config) string {

	protocol := `http`

	if config.GetBool(`minterApi.isSecure`) {
		protocol += `s://`
	} else {
		protocol += `://`
	}

	return protocol + config.GetString("minterApi.link") + `:` + config.GetString("minterApi.port")
}

//Get JSON response from API
func getJson(url string, target interface{}) error {

	r, err := httpClient.Get(url)

	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// Get last block height from Minter API
func getLastBlockFromMinterAPI(config env.Config) uint {
	statusResponse := StatusResponse{}
	getJson(getApiLink(config)+`/api/status`, &statusResponse)
	u64, err := strconv.ParseUint(statusResponse.Result.LatestBlockHeight, 10, 32)
	helpers.CheckErr(err)
	return uint(u64)
}

func getLastBlockFromDB(db *gorm.DB) uint {
	var b models.Block
	db.Last(&b)
	return b.Height
}

// Store data to DB
func storeDataToDb(config env.Config, db *gorm.DB, blockHeight uint) error {
	apiLink := getApiLink(config) + `/api/block/` + fmt.Sprint(blockHeight)
	blockResponse := BlockResponse{}
	getJson(apiLink, &blockResponse)
	blockResult := blockResponse.Result

	storeBlockToDB(db, &blockResult)

	if config.GetBool(`debug`) {
		log.Printf("Block: %d; Txs: %d; Hash: %s", blockResult.Height, blockResult.TxCount, blockResponse.Result.Hash)
	}

	return nil
}

func storeBlockToDB(db *gorm.DB, blockData *BlockResult) {

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
		BlockTime:   getBlockTime(db, blockData.Height, blockData.Time),
		BlockReward: blockData.BlockReward,
	}

	if blockModel.TxCount > 0 {
		blockModel.Transactions = getTransactionModelsFromApiData(blockData)
	}

	if blockData.Events != nil {
		e := getEventsModelsFromApiData(blockData)
		blockModel.Rewards = e.Rewards
		blockModel.Slashes = e.Slashes
	}

	db.Create(&blockModel)
}

func getBlockTime(db *gorm.DB, currentBlockHeight uint, blockTime time.Time) float64 {

	if currentBlockHeight == 1 {
		return 1
	}

	var b models.Block
	db.Where("height = ?", currentBlockHeight-1).First(&b)

	result := blockTime.Sub(b.CreatedAt)
	if result < 0 {
		return 5
	}

	return result.Seconds()
}

func getTransactionModelsFromApiData(blockData *BlockResult) []models.Transaction {

	var result = make([]models.Transaction, blockData.TxCount)

	for i, tx := range blockData.Transactions {

		status := tx.Log == nil
		payload, _ := base64.StdEncoding.DecodeString(tx.Payload)

		result[i] = models.Transaction{
			Hash:                 strings.Title(tx.Hash),
			From:                 strings.Title(tx.From),
			Type:                 tx.Type,
			Nonce:                tx.Nonce,
			GasPrice:             tx.GasPrice,
			Fee:                  tx.Gas,
			Payload:              strip.StripTags(string(payload[:])),
			ServiceData:          strip.StripTags(tx.ServiceData),
			CreatedAt:            blockData.Time,
			To:                   tx.Data.To,
			Address:              tx.Data.Address,
			Name:                 stripHtmlTags(tx.Data.Name),
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
			for k, tg := range tx.Tags {
				tags = append(tags, models.TxTag{
					Key:   k,
					Value: tg,
				})
			}
			result[i].Tags = tags
		}

		if tx.Type == models.TX_TYPE_CREATE_COIN {
			result[i].Coin = tx.Data.Symbol
		}

		if tx.Type == models.TX_TYPE_SELL_COIN {
			result[i].ValueToSell = tx.Data.ValueToSell
			result[i].ValueToBuy = getValueFromTxTag(tags, `tx.return`)
		}
		if tx.Type == models.TX_TYPE_SELL_ALL_COIN {
			result[i].ValueToSell = getValueFromTxTag(tags, `tx.sell_amount`)
			result[i].ValueToBuy = getValueFromTxTag(tags, `tx.return`)
		}
		if tx.Type == models.TX_TYPE_BUY_COIN {
			result[i].ValueToSell = getValueFromTxTag(tags, `tx.return`)
			result[i].ValueToBuy = tx.Data.ValueToBuy
		}
	}

	return result
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

func getEventsModelsFromApiData(blockData *BlockResult) events {

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
