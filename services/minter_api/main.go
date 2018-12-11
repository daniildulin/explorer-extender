package minter_api

import (
	"encoding/json"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
	"strings"
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

func (api *MinterApi) GetLastBlock() (uint64, error) {
	var err error
	response := StatusResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/status`
		err = api.getJson(link, &response)
		u64, err := strconv.ParseUint(response.Result.LatestBlockHeight, 10, 32)
		if err == nil && response.Log == nil {
			return u64, nil
		}
	}
	return 0, err
}

func (api *MinterApi) GetBlock(blockHeight uint64) (*BlockResponse, error) {
	var err error
	response := BlockResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/block?height=` + fmt.Sprint(blockHeight)
		err = api.getJson(link, &response)
		helpers.CheckErr(err)
		if response.Log == nil {
			return &response, nil
		}
	}
	return nil, err
}

func (api *MinterApi) GetBlockValidators(blockHeight uint64) (*ValidatorsResponse, error) {
	var err error
	response := ValidatorsResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/validators?height=` + fmt.Sprint(blockHeight)
		api.getJson(link, &response)
		if err == nil && response.Log == nil {
			return &response, nil
		}
	}
	return nil, err
}

func (api *MinterApi) GetCandidateInfo(pubKey string) (*CandidateResponse, error) {
	var err error
	response := CandidateResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/candidate?pubkey=` + pubKey
		err = api.getJson(link, &response)
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
		link := node.GetFullLink() + `/coinInfo/` + coin
		api.getJson(link, &response)
		if err == nil && response.Log == nil {
			return &response, nil
		}
	}
	return &response, err
}

func (api *MinterApi) GetAddressBalance(address string) (*BalanceResponse, error) {
	var err error
	response := BalanceResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/address?address=` + strings.Title(address)
		api.getJson(link, &response)
		if err == nil && response.Log == nil {
			return &response, nil
		}
	}
	return &response, err
}

func (api *MinterApi) GetBlockEvents(blockHeight uint64) (*EventsResponse, error) {
	var err error
	response := EventsResponse{}
	api.checkNodes()
	for _, node := range api.nodes {
		link := node.GetFullLink() + `/events?height=` + fmt.Sprint(blockHeight)
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
