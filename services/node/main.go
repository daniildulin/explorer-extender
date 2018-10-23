package node

import (
	"encoding/json"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 1 * time.Second}

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
	return `http://` + config.GetString("minterApi")
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
	blockResponse := blockResponse{}
	getJson(apiLink, &blockResponse)
	blockResult := blockResponse.Result

	storeBlockToDB(db, &blockResult)

	if config.GetBool(`debug`) {
		log.Printf("Block: %d; Txs: %d; Hash: %s", blockResult.Height, blockResult.TxCount, blockResponse.Result.Hash)
	}

	return nil
}

func storeBlockToDB(db *gorm.DB, blockData *blockResult) {

	if blockData.Height <= 0 {
		return
		log.Printf("Block: %d; Txs: %d; Hash: %s", blockData.Height, blockData.TxCount, blockData.Hash)
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
