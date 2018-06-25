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
	Version = "0.2.0"

	EnvProducerRouterAddress             = "BARITO_PRODUCER_ROUTER"
	EnvXtailRouterAddress                = "BARITO_XTAIL_ROUTER"
	EnvKibanaRouterAddress               = "BARITO_KIBANA_ROUTER"
	EnvBaritoMarketUrl                   = "BARITO_MARKET_URL"
	EnvBaritoProfileApiPath              = "BARITO_PROFILE_API_PATH"
	EnvBaritoProfileApiByClusternamePath = "BARITO_PROFILE_API_BY_CLUSTERNAME_PATH"

	DefaultProducerRouterAddress             = ":8081"
	DefaultXtailRouterAddress                = ":8082"
	DefaultKibanaRouterAddress               = ":8083"
	DefaultBaritoMarketUrl                   = "http://localhost:3000"
	DefaultBaritoProfileApiPath              = "/api/profile"
	DefaultBaritoProfileApiByClusternamePath = "/api/profile_by_cluster_name"
)

var (
	routerAddress               string
	kibanaRouterAddress         string
	xtailRouterAddress          string
	baritoMarketUrl             string
	profileApiPath              string
	profileApiByClusternamePath string
)

func main() {
	routerAddress = envkit.GetString(EnvProducerRouterAddress, DefaultProducerRouterAddress)
	kibanaRouterAddress = envkit.GetString(EnvKibanaRouterAddress, DefaultXtailRouterAddress)
	xtailRouterAddress = envkit.GetString(EnvXtailRouterAddress, DefaultKibanaRouterAddress)
	baritoMarketUrl = envkit.GetString(EnvBaritoMarketUrl, DefaultBaritoMarketUrl)
	profileApiPath = envkit.GetString(EnvBaritoProfileApiPath, DefaultBaritoProfileApiPath)
	profileApiByClusternamePath = envkit.GetString(EnvBaritoProfileApiByClusternamePath, DefaultBaritoProfileApiByClusternamePath)

	fmt.Printf("%s=%s\n", EnvProducerRouterAddress, routerAddress)
	fmt.Printf("%s=%s\n", EnvKibanaRouterAddress, kibanaRouterAddress)
	fmt.Printf("%s=%s\n", EnvXtailRouterAddress, xtailRouterAddress)
	fmt.Printf("%s=%s\n", EnvBaritoMarketUrl, baritoMarketUrl)
	fmt.Printf("%s=%s\n", EnvBaritoProfileApiPath, profileApiPath)
	fmt.Printf("%s=%s\n\n", EnvBaritoProfileApiByClusternamePath, profileApiByClusternamePath)

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
