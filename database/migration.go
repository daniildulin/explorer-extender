package database

import (
	"fmt"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/jinzhu/gorm"
)

func Migrate(db *gorm.DB) {
	// Use GORM automigrate for models
	fmt.Println(`Automigrate database schema.`)
	db.AutoMigrate(
		&models.Block{},
		&models.Transaction{},
		&models.TxTag{},
		&models.Reward{},
		&models.Slash{},
		&models.Validator{},
	)

	db.Exec("CREATE TABLE IF NOT EXISTS block_validator (block_id INT REFERENCES blocks (id) ON DELETE CASCADE, validator_id INT REFERENCES validators (id) ON DELETE CASCADE)")

	db.Model(&models.Transaction{}).AddForeignKey("block_id", "blocks(id)", "CASCADE", "RESTRICT")
	db.Model(&models.Reward{}).AddForeignKey("block_id", "blocks(id)", "CASCADE", "RESTRICT")
	db.Model(&models.Slash{}).AddForeignKey("block_id", "blocks(id)", "CASCADE", "RESTRICT")
	db.Model(&models.TxTag{}).AddForeignKey("transaction_id", "transaction(id)", "CASCADE", "RESTRICT")

	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_day_index ON blocks (date_trunc('day', created_at at time zone 'UTC'));`)
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_hour_index ON blocks (date_trunc('hour', created_at at time zone 'UTC'));`)
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_minute_index ON blocks (date_trunc('minute', created_at at time zone 'UTC'));`)

	db.Exec(`CREATE INDEX IF NOT EXISTS  transactions_from_index ON transactions ("from" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS  transactions_to_index ON transactions ("to" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS  transactions_hash_index ON transactions ("hash" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS  transactions_pub_key_index ON transactions ("pub_key" ASC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS  transactions_address_index ON transactions ("address" ASC)`)
}
