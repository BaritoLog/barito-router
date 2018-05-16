package main

import (
	"github.com/BaritoLog/barito-router/router"
	"os"
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
		baritoMarketUrl = "http://localhost:3000/apps"
	}

	t := router.NewTrader(baritoMarketUrl)
	r := router.NewRouter(routerAddress, t)
	r.Server().ListenAndServe()
}
