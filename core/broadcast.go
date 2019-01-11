package core

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/daniildulin/explorer-extender/models"
	"log"
	"net/http"
)

type MinterBroadcastService struct {
	httpClient *http.Client
	client     *gocent.Client
	ctx        context.Context
}

func NewMinterBroadcast(c *gocent.Client, h *http.Client) *MinterBroadcastService {
	var mbs = &MinterBroadcastService{
		httpClient: h,
		client:     c,
		ctx:        context.Background(),
	}

	return mbs
}

type StatusPageData struct {
	Data struct {
		Status              string  `json:"status"`
		Uptime              int     `json:"uptime"`
		NumberOfBlocks      int     `json:"numberOfBlocks"`
		BlockSpeed24H       float64 `json:"blockSpeed24h"`
		TxTotalCount        int     `json:"txTotalCount"`
		Tx24HCount          int     `json:"tx24hCount"`
		TxPerSecond         float64 `json:"txPerSecond"`
		ActiveValidators    int     `json:"activeValidators"`
		ActiveCandidates    int     `json:"activeCandidates"`
		AverageTxCommission float64 `json:"averageTxCommission"`
		TotalCommission     float64 `json:"totalCommission"`
	} `json:"data"`
}

func (mbs *MinterBroadcastService) Block(b *models.Block) {
	ch := `blocks`
	msg, err := json.Marshal(b.GetResponse())
	if err != nil {
		log.Printf(`Error parse json: %s`, err)
	}
	mbs.publish(ch, []byte(msg))
}

func (mbs *MinterBroadcastService) Transaction(tx *models.Transaction) {
	ch := `transactions`
	if tx.Status {
		msg, err := json.Marshal(tx.GetResponse())
		if err != nil {
			log.Printf(`Error parse json: %s`, err)
		}
		mbs.publish(ch, msg)
	}
}

func (mbs *MinterBroadcastService) Balance(b *models.Balance) {
	msg, _ := json.Marshal(b.GetResponse())
	mbs.publish(b.Address, []byte(msg))
}

func (mbs *MinterBroadcastService) StatusPage() {
	response := StatusPageData{}
	link := `https://testnet.explorer.minter.network/api/v1/status-page`
	err := mbs.getJson(link, &response)
	if err != nil {
		log.Println(err)
	}
	msg, _ := json.Marshal(response.Data)
	mbs.publish(`status_page`, []byte(msg))
}

func (mbs *MinterBroadcastService) publish(ch string, msg []byte) {
	err := mbs.client.Publish(mbs.ctx, ch, msg)
	if err != nil {
		log.Printf(`Error calling publish: %s`, err)
	}
}

func (mbs *MinterBroadcastService) getJson(url string, target interface{}) error {
	r, err := mbs.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
