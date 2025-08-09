package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	pb "github.com/bentol/barito-proto/producer"
	"github.com/mostynb/go-grpc-compression/zstd"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

const (
	AppSecretHeaderName        = "X-App-Secret"
	AppGroupSecretHeaderName   = "X-App-Group-Secret"
	RouterForwardingHeaderName = "X-Barito-Router-Forwarded"
	AppNameHeaderName          = "X-App-Name"
	KeyProducer                = "producer"
	AppNoProfilePath           = "api/producer_no_profile"
	AppNoSecretPath            = "api/no_secret"

	ErrorDoubleRouterForward = "Request already forwarded from another router, skipping forwarding again."
)

var (
	ProducerTracer = otel.Tracer("barito-router.producer")
)

type ProducerRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type producerRouter struct {
	addr                              string
	marketUrl                         string
	profilePath                       string
	profileByAppGroupPath             string
	cacheBag                          *cache.Cache
	client                            *http.Client
	appCtx                            *appcontext.AppContext
	producerStore                     *ProducerStore
	isRouterLocationForwardingEnabled bool
}

func NewProducerRouter(addr, marketUrl, profilePath string, profileByAppGroupPath string, appCtx *appcontext.AppContext) ProducerRouter {
	return &producerRouter{
		addr:                              addr,
		marketUrl:                         marketUrl,
		profilePath:                       profilePath,
		profileByAppGroupPath:             profileByAppGroupPath,
		cacheBag:                          cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:                            createClient(),
		appCtx:                            appCtx,
		producerStore:                     NewProducerStore(),
		isRouterLocationForwardingEnabled: len(config.RouterLocationForwardingMap) > 0,
	}
}

func (p *producerRouter) Server() *http.Server {
	handler := otelhttp.NewHandler(p, "/")
	return &http.Server{
		Addr:    p.addr,
		Handler: handler,
	}
}

func (p *producerRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqBody := []byte{}

	if req.URL.Path == "/ping" {
		OnPing(w, req)
		return
	}
	ctx, otelSpan := ProducerTracer.Start(req.Context(), "ServeHTTP")
	defer otelSpan.End()

	profile, err := p.getProfile(ctx, w, req)
	if p.isProfileError(w, req, profile, err, otelSpan) {
		return
	}

	instrumentation.IncreaseProducerRequestCount(
		profile.ClusterName,
		req.Header.Get(AppNameHeaderName),
		profile.ProducerAddress,
	)

	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
	}

	// check if the router enable routerLocationForwarding
	if p.isRouterLocationForwardingEnabled {
		// check if the router is eligible for forwarding, based on producer_location & this router env variables
		if host, isEligible := p.isEligibleForRouterLocationForwarding(profile); isEligible {
			// make sure the request is not forwarded from another router
			if req.Header.Get(RouterForwardingHeaderName) != "" {
				p.onDoubleRouterForward(w, req, profile.ClusterName, otelSpan)
				return
			}

			p.onEligibleForwarding(ctx, w, req, host, profile.ClusterName, reqBody)
			return
		}
	}

	//Collect all results and errors for all the produce endpoints.
	var produceResults []*pb.ProduceResult
	var produceErrors []error

	// profile.ProducerAddress is used only for the K8s infrastructure.
	if profile.ProducerAddress != "" {
		pAttrK8s := p.fetchK8sProducerAttributes(profile)
		result, err := p.handleProduce(ctx, req, reqBody, pAttrK8s, profile)

		produceResults = append(produceResults, result)
		produceErrors = append(produceErrors, err)
	}

	checkProduceResultsAndRespond(w, produceResults, produceErrors)
}

func isAppSecretAvailable(appSecret string) bool {
	if appSecret != "" {
		return true
	}
	return false
}

func isAppGroupSecretAvailable(appGroupSecret string, appName string) bool {
	if appGroupSecret != "" && appName != "" {
		return true
	}
	return false
}

func (p *producerRouter) getProfile(ctx context.Context, w http.ResponseWriter, req *http.Request) (*Profile, error) {
	var err error
	var profile *Profile

	ctx, span := ProducerTracer.Start(ctx, "getProfile")
	defer span.End()

	appSecret := req.Header.Get(AppSecretHeaderName)
	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)

	if isAppSecretAvailable(appSecret) {
		profile, err = fetchProfileByAppSecret(ctx, p.client, p.cacheBag, p.marketUrl, p.profilePath, appSecret)
		if err != nil {
			span.SetStatus(codes.Error, "Failed to fetch profile by app secret")
		}
		return profile, err
	}

	if isAppGroupSecretAvailable(appGroupSecret, appName) {
		profile, err = fetchProfileByAppGroupSecret(ctx, p.client, p.cacheBag, p.marketUrl, p.profileByAppGroupPath, appGroupSecret, appName)
		if err != nil {
			span.SetStatus(codes.Error, "Failed to fetch profile by app secret")
		}
		return profile, err

	}

	span.SetStatus(codes.Unset, "No app secret or app group secret provided")
	onNoSecret(w)
	return nil, nil
}

func (p *producerRouter) isProfileError(w http.ResponseWriter, req *http.Request, profile *Profile, err error, span trace.Span) bool {

	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)
	if err != nil {
		onTradeError(w, err)
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, profile.ProducerAddress, req, err, span)
		return true
	}

	if profile == nil {
		onNoProfile(w)
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, "", req, err, span)
		instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoProfilePath, w, req)
		return true
	}
	return false
}

func (p *producerRouter) fetchK8sProducerAttributes(profile *Profile) producerAttributes {

	producerName, _ := profile.MetaServiceName(KeyProducer)

	pAttr := producerAttributes{
		consulAddr:          "",
		producerAddr:        profile.ProducerAddress,
		producerMtlsEnabled: profile.ProducerMtlsEnabled,
		producerName:        producerName,
		appSecret:           profile.AppSecret,
	}
	return pAttr
}

func (p *producerRouter) handleProduce(ctx context.Context, req *http.Request, reqBody []byte, pAttr producerAttributes, profile *Profile) (*pb.ProduceResult, error) {
	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)
	producerClient := p.producerStore.GetClient(pAttr)

	ctx, span := ProducerTracer.Start(ctx, "handleProduce", trace.WithAttributes(
		attribute.String("appGroupSecret", appGroupSecret),
		attribute.String("appName", appName),
	))
	defer span.End()

	timberContext := TimberContextFromProfile(profile)
	var result *pb.ProduceResult

	var grpcCallOption []grpc.CallOption
	grpcCallOption = append(grpcCallOption, grpc.UseCompressor(zstd.Name))

	// Check if the request has a "Content-Encoding" header with value "gzip"
	if req.Header.Get("Content-Encoding") == "gzip" {
		// Decompress the gzip-encoded request body
		gzipReader, err := gzip.NewReader(bytes.NewReader(reqBody))
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorGzipDecompression, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err, span)
			return nil, err
		}
		defer gzipReader.Close()

		// Read the decompressed request body into a buffer
		reqBody, err = io.ReadAll(gzipReader)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorGzipDecompression, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err, span)
			return nil, err
		}
	}

	if req.URL.Path == "/produce_batch" {
		timberCollection, err := ConvertBytesToTimberCollection(reqBody, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err, span)
			return nil, err
		}

		startTime := time.Now()
		result, err = producerClient.ProduceBatch(ctx, &timberCollection, grpcCallOption...)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, pAttr.producerAddr, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err, span)
			return nil, err
		}
		instrumentation.ObserveByteIngestion(profile.ClusterName, appName, pAttr.producerAddr, reqBody)
		return result, nil

	}
	if req.URL.Path == "/produce" {
		timber, err := ConvertBytesToTimber(reqBody, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err, span)
			return nil, err
		}

		startTime := time.Now()
		result, err = producerClient.Produce(ctx, &timber, grpcCallOption...)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, pAttr.producerAddr, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err, span)
			return nil, err
		}
		instrumentation.ObserveByteIngestion(profile.ClusterName, appName, pAttr.producerAddr, reqBody)
		return result, nil
	}

	return nil, fmt.Errorf("Invalid URL called - %s", req.URL.Path)
}

func (p *producerRouter) isEligibleForRouterLocationForwarding(profile *Profile) (string, bool) {
	host, isEligible := config.RouterLocationForwardingMap[profile.ProducerLocation]
	return host, isEligible
}

func (p *producerRouter) forwardToOtherRouter(host string, req *http.Request, reqBody []byte, appGroupName string) (*http.Response, error) {
	// Create a new request to the other router
	newReq, err := http.NewRequest(req.Method, host+req.URL.Path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Copy headers from the original request to the new request
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}
	newReq.Header.Set(RouterForwardingHeaderName, "1")
	newReq.Header.Set("Accept-Encoding", "")

	// Send the request to the other router
	return p.client.Do(newReq)
}

func (p *producerRouter) onDoubleRouterForward(w http.ResponseWriter, req *http.Request, clusterName string, span trace.Span) {
	slog.Error(ErrorDoubleRouterForward)
	span.SetStatus(codes.Error, ErrorDoubleRouterForward)
	instrumentation.IncreaseDoubleRouterForward(clusterName, req.Header.Get(AppNameHeaderName))
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(ErrorDoubleRouterForward))
}

func (p *producerRouter) onEligibleForwarding(ctx context.Context, w http.ResponseWriter, req *http.Request, otherRouterHost, clusterName string, reqBody []byte) {
	_, span := ProducerTracer.Start(ctx, "onEligibleForwarding",
		trace.WithAttributes(
			attribute.String("otherRouterHost", otherRouterHost),
			attribute.String("clusterName", clusterName),
			attribute.String("appName", req.Header.Get(AppNameHeaderName)),
		),
	)
	defer span.End()

	appName := req.Header.Get(AppNameHeaderName)
	resp, err := p.forwardToOtherRouter(otherRouterHost, req, reqBody, clusterName)
	if err != nil {
		span.SetStatus(codes.Error, "Error forwarding to other router")
		slog.Error("Error forwarding to other router", slog.String("error", err.Error()))
		instrumentation.IncreaseForwardToOtherRouterFailed(clusterName, appName, otherRouterHost)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Error("Error copying response body", slog.String("error", err.Error()))
		instrumentation.IncreaseForwardToOtherRouterFailed(clusterName, appName, otherRouterHost)
		span.SetStatus(codes.Error, "Error copying response body")
		return
	}
	instrumentation.IncreaseForwardToOtherRouterSuccess(clusterName, appName, otherRouterHost)
}

func checkProduceResultsAndRespond(w http.ResponseWriter, results []*pb.ProduceResult, errors []error) {
	var validErrors []error

	// Collect all errors
	for _, err := range errors {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}

	if len(validErrors) > 0 {
		onRpcError(w, validErrors)
	} else {
		var responseMsg bytes.Buffer
		for _, result := range results {
			if result != nil {
				responseMsg.WriteString(result.Topic + "\n")
			}
		}
		onRpcSuccess(w, responseMsg.String())
	}
}
