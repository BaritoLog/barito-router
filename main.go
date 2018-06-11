package main

import (
	"fmt"

	"github.com/BaritoLog/barito-router/router"
	"github.com/BaritoLog/go-boilerplate/envkit"
)

const (
	EnvRouterAddress                     = "BARITO_PRODUCER_ROUTER"
	EnvXtailRouterAddress                = "BARITO_XTAIL_ROUTER"
	EnvKibanaRouterAddress               = "BARITO_KIBANA_ROUTER"
	EnvBaritoMarketUrl                   = "BARITO_MARKET_URL"
	EnvBaritoProfileApiPath              = "BARITO_PROFILE_API_PATH"
	EnvBaritoProfileApiByClusternamePath = "BARITO_PROFILE_API_BY_CLUSTERNAME_PATH"
	Version                              = "0.1.2"
)

func main() {
	routerAddress := envkit.GetString(EnvRouterAddress, ":8081")
	kibanaRouterAddress := envkit.GetString(EnvKibanaRouterAddress, ":8082")
	xtailRouterAddress := envkit.GetString(EnvXtailRouterAddress, ":8083")
	baritoMarketUrl := envkit.GetString(EnvBaritoMarketUrl, "http://localhost:3000")
	profileApiPath := envkit.GetString(EnvBaritoProfileApiPath, "/api/profile")
	profileApiByClusternamePath := envkit.GetString(EnvBaritoProfileApiPath, "/api/profile_by_cluster_name")

	fmt.Printf(".: Barito Router v%s :.\n\n", Version)
	fmt.Printf("%s=%s\n", EnvRouterAddress, routerAddress)
	fmt.Printf("%s=%s\n", EnvKibanaRouterAddress, kibanaRouterAddress)
	fmt.Printf("%s=%s\n", EnvXtailRouterAddress, xtailRouterAddress)
	fmt.Printf("%s=%s\n", EnvBaritoMarketUrl, baritoMarketUrl)
	fmt.Printf("%s=%s\n", EnvBaritoProfileApiPath, profileApiPath)
	fmt.Printf("%s=%s\n", EnvBaritoProfileApiByClusternamePath, profileApiByClusternamePath)

	fmt.Printf("\n")

	consul := router.NewConsulHandler()

	go func() {
		trader := router.NewTraderBySecret(baritoMarketUrl + profileApiPath)
		produceRouter := router.NewProduceRouter(routerAddress, trader, consul)
		produceRouter.Server().ListenAndServe()
	}()

	go func() {
		trader := router.NewTraderByClusterName(baritoMarketUrl + profileApiByClusternamePath)
		xtailRouter := router.NewXtailRouter(xtailRouterAddress, trader, consul)
		xtailRouter.Server().ListenAndServe()
	}()

	trader := router.NewTraderByClusterName(baritoMarketUrl + profileApiByClusternamePath)
	kibanaRouter := router.NewKibanaRouter(kibanaRouterAddress, trader, consul)
	kibanaRouter.Server().ListenAndServe()
}
