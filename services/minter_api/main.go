package minter_api

import (
	"encoding/json"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
)

type MinterApi struct {
	config     env.Config
	db         *gorm.DB
	nodes      []models.MinterNode
	httpClient *http.Client
}

func New(config env.Config, db *gorm.DB, httpClient *http.Client) *MinterApi {
	api := &MinterApi{
		config:     config,
		db:         db,
		httpClient: httpClient,
	}
	api.GetActualNodes()
	return api
}

func (api *MinterApi) GetActiveNodesCount() int {
	return len(api.nodes)
}

func (api *MinterApi) GetActualNodes() {
	var nodes []models.MinterNode
	api.db.Where("is_excluded <> ? AND is_active = ?", true, true).Order("ping asc").Find(&nodes)
	api.nodes = nodes
}

func (api *MinterApi) GetLastBlock() (uint, error) {
	var err error
	response := StatusResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/api/status`
		api.getJson(link, &response)
		u64, err := strconv.ParseUint(response.Result.LatestBlockHeight, 10, 32)
		if err == nil && response.Log == nil {
			return uint(u64), nil
		}
	}
	return 0, err
}

func (api *MinterApi) GetBlock(blockHeight uint) (*BlockResponse, error) {
	var err error
	response := BlockResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/api/block/` + fmt.Sprint(blockHeight)
		api.getJson(link, &response)
		if response.Log == nil {
			return &response, nil
		}
	}
	return nil, err
}

func (api *MinterApi) GetBlockValidators(blockHeight uint) (*ValidatorsResponse, error) {
	var err error
	response := ValidatorsResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/api/validators/?height=` + fmt.Sprint(blockHeight)
		api.getJson(link, &response)
		if err == nil && response.Log == nil {
			return &response, nil
		}
	}
	return nil, err
}

func (api *MinterApi) GetCoinInfo(coin string) (*CoinInfoResponse, error) {
	var err error
	response := CoinInfoResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/api/coinInfo/` + coin
		api.getJson(link, &response)
		if err == nil && response.Log == nil {
			return &response, nil
		}
	}
	return &response, err
}

func (api *MinterApi) getJson(url string, target interface{}) error {
	r, err := api.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func (api *MinterApi) checkNodes() {
	if len(api.nodes) == 0 {
		api.GetActualNodes()
	}
}
