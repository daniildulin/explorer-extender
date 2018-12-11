package minter_service

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/daniildulin/explorer-extender/models"
	"log"
)

type MinterBroadcastService struct {
	client *gocent.Client
	ctx    context.Context
}

func NewMinterBroadcast(c *gocent.Client) *MinterBroadcastService {
	var mbs = &MinterBroadcastService{
		client: c,
		ctx:    context.Background(),
	}

	return mbs
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

func (mbs *MinterBroadcastService) publish(ch string, msg []byte) {
	err := mbs.client.Publish(mbs.ctx, ch, msg)
	if err != nil {
		log.Printf(`Error calling publish: %s`, err)
	}
}
