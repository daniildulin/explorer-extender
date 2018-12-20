package metrics

import (
	"github.com/daniildulin/explorer-extender/env"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func Run(config env.Config) {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(config.GetString(`extenderApi.link`)+`:`+config.GetString(`extenderApi.port`), nil))
}
