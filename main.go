package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/BaritoLog/barito-router/config"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/uber/jaeger-lib/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/BaritoLog/barito-router/appcontext"
	newrelic "github.com/newrelic/go-agent"
	"github.com/urfave/cli"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func printConfig() {
	fmt.Printf("%s=%s\n", config.EnvProducerRouterAddress, config.RouterAddress)
	fmt.Printf("%s=%s\n", config.EnvKibanaRouterAddress, config.KibanaRouterAddress)
	fmt.Printf("%s=%s\n", config.EnvProducerPort, config.ProducerPort)
	fmt.Printf("%s=%s\n", config.EnvBaritoMarketUrl, config.BaritoMarketUrl)
	fmt.Printf("%s=%s\n", config.EnvBaritoViewerUrl, config.BaritoViewerUrl)
	fmt.Printf("%s=%s\n", config.EnvBaritoProfileApiPath, config.ProfileApiPath)
	fmt.Printf("%s=%s\n", config.EnvBaritoProfileApiByAppGroupPath, config.ProfileApiByAppGroupPath)
	fmt.Printf("%s=%s\n", config.EnvBaritoAuthorizeApiPath, config.AuthorizeApiPath)
	fmt.Printf("%s=%s\n", config.EnvBaritoProfileApiByClusternamePath, config.ProfileApiByClusternamePath)
	fmt.Printf("%s=%s\n", config.EnvNewRelicAppName, config.NewRelicAppName)
	fmt.Printf("%s=%s\n", config.EnvNewRelicEnabled, config.NewRelicEnabled)
	fmt.Printf("%s=%s\n", config.EnvCacheExpirationTimeSeconds, config.CacheExpirationTimeSeconds)
	fmt.Printf("%s=%s\n", config.EnvBackupCacheExpirationTimeHours, config.BackupCacheExpirationTimeHours)
	fmt.Printf("%s=%s\n", config.EnvEnableTracing, config.EnableTracing)
	fmt.Printf("%s=%s\n", config.EnvEnableSSO, config.EnableSSO)
	fmt.Printf("%s=%s\n", config.EnvSSORedirectPath, config.SSORedirectPath)
	fmt.Printf("%s=%s\n", config.EnvAllowedDomains, config.AllowedDomains)
	fmt.Printf("%s=%s\n", config.EnvRouterLocationForwardingMap, config.RouterLocationForwardingMap)
	fmt.Printf("%s=%s\n", config.EnvJaegerServiceName, config.JaegerServiceName)
}

func main() {
	printConfig()

	newRelicConfig := newrelic.NewConfig(config.NewRelicAppName, config.NewRelicLicenseKey)
	newRelicConfig.Enabled = config.NewRelicEnabled
	appCtx := appcontext.NewAppContext(newRelicConfig)

	// enable tracing
	if config.EnableTracing {
		// Handle SIGINT (CTRL+C) gracefully.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		// Set up OpenTelemetry.
		otelShutdown, err := setupOTelSDK(ctx)
		if err != nil {
			return
		}
		// Handle shutdown properly so nothing leaks.
		defer func() {
			err = errors.Join(err, otelShutdown(context.Background()))
		}()
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

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider() (*trace.TracerProvider, error) {
	traceExporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint("localhost:4317"),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return tracerProvider, nil
}
