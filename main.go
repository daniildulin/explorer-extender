package main

import (
	"flag"
	"fmt"
	"github.com/centrifugal/gocent"
	"github.com/daniildulin/explorer-extender/core"
	"github.com/daniildulin/explorer-extender/database"
	"github.com/daniildulin/explorer-extender/env"
	"github.com/daniildulin/explorer-extender/helpers"
	"github.com/daniildulin/explorer-extender/metrics"
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

	go metrics.Run(config)

	db, err := gorm.Open("postgres", config.GetString(`database.url`))
	helpers.CheckErr(err)
	defer db.Close()
	db.LogMode(config.GetBool(`debug`))
	database.Migrate(db, config)

	httpClient := &http.Client{Timeout: 30 * time.Second}

	wsLink := `http://`
	if config.GetBool(`wsServer.isSecure`) {
		wsLink = `https://`
	}

	wsLink += config.GetString(`wsServer.link`)

	if config.GetString(`wsServer.port`) != `` {
		wsLink += `:` + config.GetString(`wsServer.port`)
	}

	wsClient := gocent.New(gocent.Config{
		Addr:       wsLink,
		Key:        config.GetString(`wsServer.key`),
		HTTPClient: httpClient,
	})

	mbs := core.NewMinterBroadcast(wsClient, httpClient)
	minterService := core.New(config, db, mbs)
	minterService.Run()
}
