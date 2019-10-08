package main

import (
	"fmt"
	"github.com/BaritoLog/barito-router/config"
	"log"
	"os"

	"github.com/BaritoLog/barito-router/appcontext"
	newrelic "github.com/newrelic/go-agent"
	"github.com/urfave/cli"
)

func main() {
	fmt.Printf("%s=%s\n", config.EnvProducerRouterAddress, config.RouterAddress)
	fmt.Printf("%s=%s\n", config.EnvKibanaRouterAddress, config.KibanaRouterAddress)
	fmt.Printf("%s=%s\n", config.EnvBaritoMarketUrl, config.BaritoMarketUrl)
	fmt.Printf("%s=%s\n", config.EnvBaritoMarketAccessToken, config.BaritoMarketAccessToken)
	fmt.Printf("%s=%s\n", config.EnvBaritoProfileApiPath, config.ProfileApiPath)
	fmt.Printf("%s=%s\n", config.EnvBaritoProfileApiByAppGroupPath, config.ProfileApiByAppGroupPath)
	fmt.Printf("%s=%s\n", config.EnvBaritoAuthorizeApiPath, config.AuthorizeApiPath)
	fmt.Printf("%s=%s\n\n", config.EnvBaritoProfileApiByClusternamePath, config.ProfileApiByClusternamePath)
	fmt.Printf("%s=%s\n", config.EnvCASAddress, config.CasAddress)

	newRelicConfig := newrelic.NewConfig(config.NewRelicAppName, config.NewRelicLicenseKey)
	newRelicConfig.Enabled = config.NewRelicEnabled
	appCtx := appcontext.NewAppContext(newRelicConfig)

	app := cli.App{
		Name:    config.Name,
		Usage:   "Route from outside world to barito world",
		Version: config.Version,
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
