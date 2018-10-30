package coins

import (
	"github.com/daniildulin/explorer-extender/models"
	"github.com/jinzhu/gorm"
)

func Store(c chan models.Coin, db *gorm.DB) {
	for {
		coin := <-c
		db.Save(&coin)
	}
}

func CreateFromTransaction(c chan models.Coin, transaction models.Transaction) {
	c <- models.Coin{
		Symbol:         *transaction.Coin,
		Name:           *transaction.Name,
		Volume:         *transaction.InitialAmount,
		ReserveBalance: *transaction.InitialReserve,
		Crr:            *transaction.ConstantReserveRatio,
		Creator:        transaction.From,
		CreatedAt:      transaction.CreatedAt,
	}
}
