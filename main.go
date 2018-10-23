package main

import (
	"flag"
	"fmt"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/models"
	"github.com/daniildulin/explorer-extender/services/node"
	"github.com/jinzhu/gorm"
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var Version string   // Version
var GitCommit string // Git commit
var BuildDate string // Build date
var AppName string   // Application name
var config env.Config

var version = flag.Bool(`v`, false, `Prints current version`)

// Initialize app.
func init() {
	config = env.NewViperConfig()
	AppName = config.GetString(`name`)
	Version = `0.1`

	if config.GetBool(`debug`) {
		fmt.Println(`Service RUN on DEBUG mode`)
	}
}

func migrate(db *gorm.DB) {
	// Use GORM automigrate for models
	fmt.Println(`Automigrate database schema.`)
	db.AutoMigrate(&models.Block{})
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_day_index ON blocks (date_trunc('day', created_at at time zone 'UTC'));`)
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_hour_index ON blocks (date_trunc('hour', created_at at time zone 'UTC'));`)
	db.Exec(`CREATE INDEX IF NOT EXISTS blocks_date_trunc_minute_index ON blocks (date_trunc('minute', created_at at time zone 'UTC'));`)
}

func main() {

	flag.Parse()

	if *version {
		fmt.Printf(`%s v%s Commit %s builded %s\n`, AppName, Version, GitCommit, BuildDate)
		os.Exit(0)
	}

	db, err := gorm.Open("postgres", config.GetString(`database.url`))
	helpers.CheckErr(err)
	defer db.Close()

	migrate(db)

	node.Run(config, db)
}
