package main

import (
	"flag"
	"fmt"
	"github.com/daniildulin/explorer-extender/database"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/services/minter_api"
	"github.com/daniildulin/explorer-extender/services/minter_service"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"net/http"
	"os"
	"time"
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

func main() {
	flag.Parse()
	if *version {
		fmt.Printf(`%s v%s Commit %s builded %s\n`, AppName, Version, GitCommit, BuildDate)
		os.Exit(0)
	}

	db, err := gorm.Open("postgres", config.GetString(`database.url`))
	helpers.CheckErr(err)
	defer db.Close()
	db.LogMode(config.GetBool(`debug`))
	database.Migrate(db)

	minterApi := minter_api.New(config, db, &http.Client{Timeout: 10 * time.Second})
	minterService := minter_service.New(config, db, minterApi)

	for {
		if minterService.GetActiveNodesCount() > 0 {
			minterService.Run()
		} else {
			if config.GetBool(`debug`) {
				fmt.Println(`Waiting for available node`)
			}
			minterService.UpdateApiNodesList()
			time.Sleep(5 * time.Second)
		}
	}
}
