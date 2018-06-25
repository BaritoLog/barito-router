package main

import (
	"fmt"

	"github.com/BaritoLog/barito-router/router"
	"github.com/BaritoLog/go-boilerplate/srvkit"
	"github.com/urfave/cli"
)

func CmdKibana(ctx *cli.Context) {
	fmt.Println("::: Kibana Router :::")

	go RunKibanaRouter()
	srvkit.GracefullShutdown(func() {
		fmt.Println("Gracefull Shutdown")
	})
}

func CmdAll(cli.Context) {
	go RunProducerRouter()
	go RunKibanaRouter()
	srvkit.GracefullShutdown(func() {
		fmt.Println("Gracefull Shutdown")
	})
}

func RunProducerRouter() {
	consul := router.NewConsulHandler()
	trader := router.NewTraderBySecret(baritoMarketUrl + profileApiPath)
	produceRouter := router.NewProduceRouter(routerAddress, trader, consul)
	produceRouter.Server().ListenAndServe()

}

func RunXtailRouter() {
	consul := router.NewConsulHandler()
	trader := router.NewTraderByClusterName(baritoMarketUrl + profileApiByClusternamePath)
	xtailRouter := router.NewXtailRouter(xtailRouterAddress, trader, consul)
	xtailRouter.Server().ListenAndServe()
}

func RunKibanaRouter() {
	kibanaRouter := router.NewKibanaRouter(kibanaRouterAddress, baritoMarketUrl, profileApiByClusternamePath)
	kibanaRouter.Server().ListenAndServe()
}
