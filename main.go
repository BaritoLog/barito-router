package main

import (
	"os"

	"github.com/BaritoLog/barito-router/router"
)

const (
	EnvRouterAddress   = "BARITO_ROUTER_ADDRESS"
	EnvBaritoMarketUrl = "BARITO_ROUTER_MARKET_URL"
)

func main() {
	routerAddress := os.Getenv(EnvRouterAddress)
	if routerAddress == "" {
		routerAddress = ":8080"
	}

	baritoMarketUrl := os.Getenv(EnvBaritoMarketUrl)
	if baritoMarketUrl == "" {
		baritoMarketUrl = "http://localhost:3000/api/apps"
	}

	t := router.NewTrader(baritoMarketUrl)
	c := router.NewConsulHandler()
	r := router.NewRouter(routerAddress, t, c)
	r.Server().ListenAndServe()
}
