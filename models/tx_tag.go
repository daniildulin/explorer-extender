package models

type TxTag struct {
	ID            uint `gorm:"primary_key"`
	TransactionId uint
	Key           string
	Value         string
}
