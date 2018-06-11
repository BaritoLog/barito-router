package main

import (
	"fmt"
	"os"

	"github.com/BaritoLog/barito-router/router"
)

const (
	EnvRouterAddress       = "BARITO_ROUTER_ADDRESS"
	EnvXtailRouterAddress  = "BARITO_XTAIL_ROUTER_ADDRESS"
	EnvKibanaRouterAddress = "BARITO_KIBANA_ROUTER_ADDRESS"
	EnvBaritoMarketUrl     = "BARITO_ROUTER_MARKET_URL"
	Version                = "0.1.2"
)

func main() {
	routerAddress := os.Getenv(EnvRouterAddress)
	if routerAddress == "" {
		routerAddress = ":8081"
	}

	kibanaRouterAddress := os.Getenv(EnvKibanaRouterAddress)
	if kibanaRouterAddress == "" {
		kibanaRouterAddress = ":8082"
	}

	xtailRouterAddress := os.Getenv(EnvXtailRouterAddress)
	if xtailRouterAddress == "" {
		xtailRouterAddress = ":8083"
	}

	baritoMarketUrl := os.Getenv(EnvBaritoMarketUrl)
	if baritoMarketUrl == "" {
		baritoMarketUrl = "http://localhost:3000/api/profile"
	}

	fmt.Printf(".: Barito Router v%s :.\n\n", Version)
	fmt.Printf("%s=%s\n", EnvRouterAddress, routerAddress)
	fmt.Printf("%s=%s\n", EnvKibanaRouterAddress, kibanaRouterAddress)
	fmt.Printf("%s=%s\n", EnvXtailRouterAddress, xtailRouterAddress)
	fmt.Printf("%s=%s\n", EnvBaritoMarketUrl, baritoMarketUrl)
	fmt.Printf("\n")

	trader := router.NewTrader(baritoMarketUrl)
	consul := router.NewConsulHandler()

	go func() {
		produceRouter := router.NewProduceRouter(routerAddress, trader, consul)
		produceRouter.Server().ListenAndServe()
	}()

	go func() {
		xtailRouter := router.NewXtailRouter(xtailRouterAddress, trader, consul)
		xtailRouter.Server().ListenAndServe()
	}()

	kibanaRouter := router.NewKibanaRouter(kibanaRouterAddress, trader, consul)
	kibanaRouter.Server().ListenAndServe()
}
