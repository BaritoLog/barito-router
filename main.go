package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BaritoLog/go-boilerplate/envkit"
	"github.com/urfave/cli"
)

const (
	Name    = "Barito Router"
	Version = "0.3.4"

	EnvProducerRouterAddress             = "BARITO_PRODUCER_ROUTER"
	EnvKibanaRouterAddress               = "BARITO_KIBANA_ROUTER"
	EnvBaritoMarketUrl                   = "BARITO_MARKET_URL"
	EnvBaritoProfileApiPath              = "BARITO_PROFILE_API_PATH"
	EnvBaritoAuthorizeApiPath            = "BARITO_AUTHORIZE_API_PATH"
	EnvBaritoProfileApiByClusternamePath = "BARITO_PROFILE_API_BY_CLUSTERNAME_PATH"
	EnvCASAddress                        = "BARITO_CAS_ADDRESS"

	DefaultProducerRouterAddress             = ":8081"
	DefaultKibanaRouterAddress               = ":8083"
	DefaultBaritoMarketUrl                   = "http://localhost:3000"
	DefaultBaritoProfileApiPath              = "api/profile"
	DefaultBaritoAuthorizeApiPath            = "api/authorize"
	DefaultBaritoProfileApiByClusternamePath = "api/profile_by_cluster_name"
	DefaultCASAddress                        = ""
)

var (
	routerAddress               string
	kibanaRouterAddress         string
	baritoMarketUrl             string
	profileApiPath              string
	authorizeApiPath            string
	profileApiByClusternamePath string
	casAddress                  string
)

func main() {
	routerAddress, _ = envkit.GetString(
		EnvProducerRouterAddress,
		DefaultProducerRouterAddress,
	)
	kibanaRouterAddress, _ = envkit.GetString(
		EnvKibanaRouterAddress,
		DefaultKibanaRouterAddress,
	)
	baritoMarketUrl, _ = envkit.GetString(
		EnvBaritoMarketUrl,
		DefaultBaritoMarketUrl,
	)
	profileApiPath, _ = envkit.GetString(
		EnvBaritoProfileApiPath,
		DefaultBaritoProfileApiPath,
	)
	authorizeApiPath, _ = envkit.GetString(
		EnvBaritoAuthorizeApiPath,
		DefaultBaritoAuthorizeApiPath,
	)
	profileApiByClusternamePath, _ = envkit.GetString(
		EnvBaritoProfileApiByClusternamePath,
		DefaultBaritoProfileApiByClusternamePath,
	)
	casAddress, _ = envkit.GetString(
		EnvCASAddress,
		DefaultCASAddress,
	)

	fmt.Printf("%s=%s\n", EnvProducerRouterAddress, routerAddress)
	fmt.Printf("%s=%s\n", EnvKibanaRouterAddress, kibanaRouterAddress)
	fmt.Printf("%s=%s\n", EnvBaritoMarketUrl, baritoMarketUrl)
	fmt.Printf("%s=%s\n", EnvBaritoProfileApiPath, profileApiPath)
	fmt.Printf("%s=%s\n", EnvBaritoAuthorizeApiPath, authorizeApiPath)
	fmt.Printf("%s=%s\n\n", EnvBaritoProfileApiByClusternamePath, profileApiByClusternamePath)
	fmt.Printf("%s=%s\n", EnvCASAddress, casAddress)

	app := cli.App{
		Name:    Name,
		Usage:   "Route from outside world to barito world",
		Version: Version,
		Commands: []cli.Command{
			{
				Name:      "kibana",
				ShortName: "k",
				Usage:     "kibana router",
				Action:    CmdKibana,
			},
			{
				Name:      "producer",
				ShortName: "p",
				Usage:     "producer router",
				Action:    CmdProducer,
			},
			{
				Name:      "all",
				ShortName: "a",
				Usage:     "all router",
				Action:    CmdAll,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}
