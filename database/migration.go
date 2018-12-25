package database

import (
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/jinzhu/gorm"
)

func Migrate(db *gorm.DB, config env.Config) {
	// Use GORM automigrate for models
	fmt.Println(`Automigrate database schema.`)
	db.AutoMigrate(
		&models.Block{},
		&models.Transaction{},
		&models.TxTag{},
		&models.Reward{},
		&models.Slash{},
		&models.Validator{},
		&models.Coin{},
		&models.MinterNode{},
		&models.Balance{},
		&models.Stake{},
	)

	db.Exec("CREATE TABLE IF NOT EXISTS block_validator (block_id INT REFERENCES blocks (id) ON DELETE CASCADE, validator_id INT REFERENCES validators (id) ON DELETE CASCADE)")

	db.Model(&models.Transaction{}).AddForeignKey("block_id", "blocks(id)", "CASCADE", "RESTRICT")
	db.Model(&models.Reward{}).AddForeignKey("block_id", "blocks(id)", "CASCADE", "RESTRICT")
	db.Model(&models.Slash{}).AddForeignKey("block_id", "blocks(id)", "CASCADE", "RESTRICT")
	db.Model(&models.TxTag{}).AddForeignKey("transaction_id", "transactions(id)", "CASCADE", "RESTRICT")
	db.Model(&models.Stake{}).AddForeignKey("validator_id", "validators(id)", "CASCADE", "RESTRICT")

	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_day_index ON blocks (date_trunc('day', created_at at time zone 'UTC'));`)
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_hour_index ON blocks (date_trunc('hour', created_at at time zone 'UTC'));`)
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_minute_index ON blocks (date_trunc('minute', created_at at time zone 'UTC'));`)

	db.Exec(`CREATE INDEX IF NOT EXISTS transactions_from_index ON transactions ("from" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS transactions_to_index ON transactions ("to" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS transactions_hash_index ON transactions ("hash" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS transactions_pub_key_index ON transactions ("pub_key" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS transactions_address_index ON transactions ("address" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS balance_address_index ON balances ("address" ASC)`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS balance_address_unique_index ON balances ("address", "coin")`)
	db.Exec(`CREATE INDEX IF NOT EXISTS stake_address_index ON stakes ("owner" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS stake_coin_index ON stakes ("coin" ASC)`)

	//Add Base Coin To DB
	db.FirstOrCreate(&models.Coin{}, models.Coin{
		Name:           "Minter Coin",
		Symbol:         config.GetString(`baseCoin`),
		Crr:            0,
		Volume:         `0`,
		ReserveBalance: `0`,
		Creator:        ``,
	})

	//Add Minter Node
	createNodeFromConfig(db, config)
}

func createNodeFromConfig(db *gorm.DB, config env.Config) {
	var node models.MinterNode
	db.Where("host = ? AND port = ?", config.GetString(`minterApi.link`), config.GetString(`minterApi.port`)).First(&node)
	if node.ID == 0 {
		node.Host = config.GetString(`minterApi.link`)
		node.Port = uint(config.GetInt(`minterApi.port`))
		node.IsSecure = config.GetBool(`minterApi.isSecure`)
		node.IsLocal = true
		node.IsActive = true
		db.Create(&node)
	}
}
