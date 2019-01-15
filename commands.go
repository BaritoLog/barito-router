package main

import (
	"fmt"

	"github.com/BaritoLog/barito-router/router"
	"github.com/BaritoLog/go-boilerplate/srvkit"
	"github.com/urfave/cli"
	"github.com/BaritoLog/barito-router/appcontext"
)

func CmdKibana(ctx *cli.Context) {
	fmt.Println("::: Kibana Router :::")

	go RunKibanaRouter()
	srvkit.GracefullShutdown(func() {
		fmt.Println("Graceful Shutdown")
	})
}

func CmdProducer(appCtx *appcontext.AppContext) {
	fmt.Println("::: Producer Router :::")

	go RunProducerRouter(appCtx)
	srvkit.GracefullShutdown(func() {
		fmt.Println("Graceful Shutdown")
	})
}

func CmdAll(appCtx *appcontext.AppContext ) {
	fmt.Println("::: All Router :::")

	go RunProducerRouter(appCtx)
	go RunKibanaRouter()
	srvkit.GracefullShutdown(func() {
		fmt.Println("Graceful Shutdown")
	})
}

func RunProducerRouter(appCtx *appcontext.AppContext) {
	produceRouter := router.NewProducerRouter(
		routerAddress,
		baritoMarketUrl,
		profileApiPath,
		profileApiByAppGroupPath,
		appCtx,
	)
	produceRouter.Server().ListenAndServe()

}

func RunKibanaRouter() {
	kibanaRouter := router.NewKibanaRouter(
		kibanaRouterAddress,
		baritoMarketUrl,
		profileApiByClusternamePath,
		authorizeApiPath,
		casAddress,
	)
	kibanaRouter.Server().ListenAndServe()
}
