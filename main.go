package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/go-boilerplate/envkit"
	newrelic "github.com/newrelic/go-agent"
	"github.com/urfave/cli"
)

const (
	Name    = "Barito Router"
	Version = "0.5.2"

	EnvProducerRouterAddress             = "BARITO_PRODUCER_ROUTER"
	EnvKibanaRouterAddress               = "BARITO_KIBANA_ROUTER"
	EnvBaritoMarketUrl                   = "BARITO_MARKET_URL"
	EnvBaritoProfileApiPath              = "BARITO_PROFILE_API_PATH"
	EnvBaritoProfileApiByAppGroupPath    = "BARITO_PROFILE_API_BY_APP_GROUP_PATH"
	EnvBaritoAuthorizeApiPath            = "BARITO_AUTHORIZE_API_PATH"
	EnvBaritoProfileApiByClusternamePath = "BARITO_PROFILE_API_BY_CLUSTERNAME_PATH"
	EnvCASAddress                        = "BARITO_CAS_ADDRESS"
	EnvNewRelicAppName                   = "BARITO_NEW_RELIC_APP_NAME"
	EnvNewRelicLicenseKey                = "BARITO_NEW_RELIC_LICENSE_KEY"
	EnvNewRelicEnabled                   = "BARITO_NEW_RELIC_ENABLED"

	DefaultProducerRouterAddress             = ":8081"
	DefaultKibanaRouterAddress               = ":8083"
	DefaultBaritoMarketUrl                   = "http://localhost:3000"
	DefaultBaritoProfileApiPath              = "api/profile"
	DefaultBaritoProfileApiByAppGroupPath    = "api/profile_by_app_group"
	DefaultBaritoAuthorizeApiPath            = "api/authorize"
	DefaultBaritoProfileApiByClusternamePath = "api/profile_by_cluster_name"
	DefaultCASAddress                        = ""
	DefaultNewRelicAppName                   = "barito_router"
	DefaultNewRelicLicenseKey                = ""
	DefaultNewRelicEnabled                   = false
)

var (
	routerAddress               string
	kibanaRouterAddress         string
	baritoMarketUrl             string
	profileApiPath              string
	profileApiByAppGroupPath    string
	authorizeApiPath            string
	profileApiByClusternamePath string
	casAddress                  string
	newRelicAppName             string
	newRelicLicenseKey          string
	newRelicEnabled             bool
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
	profileApiByAppGroupPath, _ = envkit.GetString(
		EnvBaritoProfileApiByAppGroupPath,
		DefaultBaritoProfileApiByAppGroupPath,
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
	newRelicAppName, _ = envkit.GetString(
		EnvNewRelicAppName,
		DefaultNewRelicAppName,
	)
	newRelicLicenseKey, _ = envkit.GetString(
		EnvNewRelicLicenseKey,
		DefaultNewRelicLicenseKey,
	)
	newRelicEnabled, _ = envkit.GetBool(
		EnvNewRelicEnabled,
		DefaultNewRelicEnabled,
	)

	fmt.Printf("%s=%s\n", EnvProducerRouterAddress, routerAddress)
	fmt.Printf("%s=%s\n", EnvKibanaRouterAddress, kibanaRouterAddress)
	fmt.Printf("%s=%s\n", EnvBaritoMarketUrl, baritoMarketUrl)
	fmt.Printf("%s=%s\n", EnvBaritoProfileApiPath, profileApiPath)
	fmt.Printf("%s=%s\n", EnvBaritoProfileApiByAppGroupPath, profileApiByAppGroupPath)
	fmt.Printf("%s=%s\n", EnvBaritoAuthorizeApiPath, authorizeApiPath)
	fmt.Printf("%s=%s\n\n", EnvBaritoProfileApiByClusternamePath, profileApiByClusternamePath)
	fmt.Printf("%s=%s\n", EnvCASAddress, casAddress)

	newRelicConfig := newrelic.NewConfig(newRelicAppName, newRelicLicenseKey)
	newRelicConfig.Enabled = newRelicEnabled
	appCtx := appcontext.NewAppContext(newRelicConfig)

	app := cli.App{
		Name:    Name,
		Usage:   "Route from outside world to barito world",
		Version: Version,
		Commands: []cli.Command{
			{
				Name:      "kibana",
				ShortName: "k",
				Usage:     "kibana router",
				Action: func(c *cli.Context) error {
					CmdKibana(appCtx)
					return nil
				},
			},
			{
				Name:      "producer",
				ShortName: "p",
				Usage:     "producer router",
				Action: func(c *cli.Context) error {
					CmdProducer(appCtx)
					return nil
				},
			},
			{
				Name:      "all",
				ShortName: "a",
				Usage:     "all router",
				Action: func(c *cli.Context) error {
					CmdAll(appCtx)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}
