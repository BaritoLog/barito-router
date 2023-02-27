package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BaritoLog/barito-router/config"
	"github.com/gorilla/mux"

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
	ssoClient := router.NewSSOClient(config.SSOClientID, config.SSOClientSecret, config.BaritoViewerUrl+config.SSORedirectPath, config.AllowedDomains, config.HMACJWTSecretString)
	kibanaRouter := router.NewKibanaRouterWithSSO(
		config.KibanaRouterAddress,
		config.BaritoMarketUrl,
		config.BaritoMarketAccessToken,
		config.ProfileApiByClusternamePath,
		config.AuthorizeApiPath,
		appCtx,
		*ssoClient,
	)

	r := mux.NewRouter()
	r.HandleFunc("/ping", router.OnPing)

	r.HandleFunc(router.PATH_LOGIN, ssoClient.HandleLogin)
	r.HandleFunc(router.PATH_CALLBACK, ssoClient.HandleCallback)

	kibanaRoute := r.PathPrefix("/").Subrouter()
	kibanaRoute.PathPrefix("/{cluster_name}").Handler(kibanaRouter)
	kibanaRoute.Use(ssoClient.MustBeAuthenticatedMiddleware)
	kibanaRoute.Use(kibanaRouter.MustBeAuthorizedMiddleware)

	srv := &http.Server{
		Handler:      r,
		Addr:         config.KibanaRouterAddress,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	srv.ListenAndServe()
}
