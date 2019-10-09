package main

import (
	"fmt"
	"github.com/BaritoLog/barito-router/config"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/router"
	"github.com/BaritoLog/go-boilerplate/srvkit"
)

func CmdKibana(appCtx *appcontext.AppContext) {
	fmt.Println("::: Kibana Router :::")

	go RunKibanaRouter(appCtx)
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

func CmdAll(appCtx *appcontext.AppContext) {
	fmt.Println("::: All Router :::")

	go RunProducerRouter(appCtx)
	go RunKibanaRouter(appCtx)
	srvkit.GracefullShutdown(func() {
		fmt.Println("Graceful Shutdown")
	})
}

func RunProducerRouter(appCtx *appcontext.AppContext) {
	produceRouter := router.NewProducerRouter(
		config.RouterAddress,
		config.BaritoMarketUrl,
		config.ProfileApiPath,
		config.ProfileApiByAppGroupPath,
		appCtx,
	)
	produceRouter.Server().ListenAndServe()

}

func RunKibanaRouter(appCtx *appcontext.AppContext) {
	kibanaRouter := router.NewKibanaRouter(
		config.KibanaRouterAddress,
		config.BaritoMarketUrl,
		config.BaritoMarketAccessToken,
		config.ProfileApiByClusternamePath,
		config.AuthorizeApiPath,
		config.CasAddress,
		appCtx,
	)
	kibanaRouter.Server().ListenAndServe()
}
