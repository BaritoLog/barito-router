package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/BaritoLog/barito-router/config"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/BaritoLog/barito-router/appcontext"
	newrelic "github.com/newrelic/go-agent"
	"github.com/urfave/cli"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
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
	fmt.Printf("%s=%v\n", config.EnvEnableTracing, config.EnableTracing)
	fmt.Printf("%s=%v\n", config.EnvEnableSSO, config.EnableSSO)

	newRelicConfig := newrelic.NewConfig(config.NewRelicAppName, config.NewRelicLicenseKey)
	newRelicConfig.Enabled = config.NewRelicEnabled
	appCtx := appcontext.NewAppContext(newRelicConfig)

	// enable tracing
	if config.EnableTracing {
		err, closer := initTracer()
		if err != nil {
			log.Fatal(fmt.Sprintf("Failed to init tracer: %s", err.Error()))
		}
		defer closer.Close()
	}

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

	http.Handle("/metrics", promhttp.Handler())
	exporterPort, exists := os.LookupEnv("EXPORTER_PORT")
	if !exists {
		exporterPort = ":8008"
	}
	go http.ListenAndServe(exporterPort, nil)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}

func initTracer() (error, io.Closer) {
	// config from environment variable
	cfg, err := jaegercfg.FromEnv()
	fmt.Printf("Jaeger cfg: %+v", cfg)
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
		return err, nil
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	// Create tracer and then initialize global tracer
	closer, err := cfg.InitGlobalTracer(
		config.JaegerServiceName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)

	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return err, nil
	}

	return nil, closer
}
