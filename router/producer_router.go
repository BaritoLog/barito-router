package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	pb "github.com/bentol/barito-proto/producer"
	"github.com/mostynb/go-grpc-compression/zstd"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
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
	return &http.Server{
		Addr:    p.addr,
		Handler: p,
	}
}

func (p *producerRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqBody := []byte{}
	if req.URL.Path == "/ping" {
		OnPing(w, req)
		return
	}

	span := opentracing.StartSpan("barito_router_producer.produce_log")
	defer span.Finish()

	profile, err := p.getProfile(w, req, span)
	if p.isProfileError(w, req, profile, err) {
		return
	}

	instrumentation.IncreaseProducerRequestCount(
		profile.ClusterName,
		req.Header.Get(AppNameHeaderName),
		profile.ProducerAddress,
	)
	span.SetTag("app-group", profile.ClusterName)

	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
	}

	// check if the router enable routerLocationForwarding
	if p.isRouterLocationForwardingEnabled {
		// make sure the request is not forwarded from another router
		if req.Header.Get(RouterForwardingHeaderName) == "1" {
			log.Infof("Request already forwarded from another router, skipping forwarding again.")
			instrumentation.IncreaseDoubleRouterForward(profile.ClusterName, req.Header.Get(AppNameHeaderName))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		host, isEligible := p.isEligibleForRouterLocationForwarding(profile)
		if isEligible {
			resp, err := p.forwardToOtherRouter(host, req, reqBody, profile.ClusterName)
			if err != nil {
				log.Errorf("Error forwarding to other router: %s", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close() // Ensure the response body is closed after reading

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("Error reading response body: %s", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(resp.StatusCode)
			w.Write(body)
			return
		}
	}

	//Collect all results and errors for all the produce endpoints.
	var produceResults []*pb.ProduceResult
	var produceErrors []error

	// ConsulHosts are only used in the legacy infrastructure(LXC containers).
	if len(profile.ConsulHosts) > 0 {
		pAttrConsul, err := p.fetchProducerAttributesFromConsul(w, req, profile)
		if err == nil {
			result, err := p.handleProduce(req, reqBody, pAttrConsul, profile)

			produceResults = append(produceResults, result)
			produceErrors = append(produceErrors, err)
		} else {
			produceErrors = append(produceErrors, err)
		}
	}

	// profile.ProducerAddress is used only for the K8s infrastructure.
	if profile.ProducerAddress != "" {
		pAttrK8s := p.fetchK8sProducerAttributes(profile)
		result, err := p.handleProduce(req, reqBody, pAttrK8s, profile)

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

func (p *producerRouter) getProfile(w http.ResponseWriter, req *http.Request, span opentracing.Span) (*Profile, error) {

	var err error
	var profile *Profile

	appSecret := req.Header.Get(AppSecretHeaderName)
	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)

	if isAppSecretAvailable(appSecret) {
		profile, err = fetchProfileByAppSecret(p.client, span.Context(), p.cacheBag, p.marketUrl, p.profilePath, appSecret)
		if profile != nil {
			instrumentation.RunTransaction(p.appCtx.NewRelicApp(), p.profileByAppGroupPath, w, req)
		}
		return profile, err
	}

	if isAppGroupSecretAvailable(appGroupSecret, appName) {
		profile, err = fetchProfileByAppGroupSecret(p.client, span.Context(), p.cacheBag, p.marketUrl, p.profileByAppGroupPath, appGroupSecret, appName)
		if profile != nil {
			instrumentation.RunTransaction(p.appCtx.NewRelicApp(), p.profileByAppGroupPath, w, req)
		}
		return profile, err

	}

	onNoSecret(w)
	instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoSecretPath, w, req)
	return nil, nil
}

func (p *producerRouter) isProfileError(w http.ResponseWriter, req *http.Request, profile *Profile, err error) bool {

	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)
	if err != nil {
		onTradeError(w, err)
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, profile.ProducerAddress, req, err)
		return true
	}

	if profile == nil {
		onNoProfile(w)
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, "", req, err)
		instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoProfilePath, w, req)
		return true
	}
	return false
}

func (p *producerRouter) fetchProducerAttributesFromConsul(w http.ResponseWriter, req *http.Request, profile *Profile) (producerAttributes, error) {

	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)
	producerName, _ := profile.MetaServiceName(KeyProducer)

	srv, consulAddr, err := consulService(profile.ConsulHosts, producerName, profile.ClusterName, p.cacheBag)
	if err != nil {
		onConsulError(w, err)
		logProduceError(instrumentation.ErrorConsulCall, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
		return producerAttributes{}, err
	}
	if srv == nil {
		err = fmt.Errorf("Can't find service from consul: %s", KeyProducer)
		onConsulError(w, err)
		logProduceError(instrumentation.ErrorNoProducer, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
		return producerAttributes{}, err
	}

	if config.ProducerPort != "" {
		port, err := strconv.Atoi(config.ProducerPort)
		if err == nil {
			srv.ServicePort = port
		}
	}

	pAttr := producerAttributes{
		consulAddr:   consulAddr,
		producerAddr: fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
		producerName: producerName,
		appSecret:    profile.AppSecret,
	}
	return pAttr, nil
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

func (p *producerRouter) handleProduce(req *http.Request, reqBody []byte, pAttr producerAttributes, profile *Profile) (*pb.ProduceResult, error) {
	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)
	producerClient := p.producerStore.GetClient(pAttr)
	ctx := context.Background()

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
			logProduceError(instrumentation.ErrorGzipDecompression, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
			return nil, err
		}
		defer gzipReader.Close()

		// Read the decompressed request body into a buffer
		reqBody, err = io.ReadAll(gzipReader)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorGzipDecompression, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
			return nil, err
		}
	}

	if req.URL.Path == "/produce_batch" {
		timberCollection, err := ConvertBytesToTimberCollection(reqBody, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
			return nil, err
		}

		startTime := time.Now()
		result, err = producerClient.ProduceBatch(ctx, &timberCollection, grpcCallOption...)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, pAttr.producerAddr, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
			return nil, err
		}
		instrumentation.ObserveByteIngestion(profile.ClusterName, appName, pAttr.producerAddr, reqBody)
		return result, nil

	}
	if req.URL.Path == "/produce" {
		timber, err := ConvertBytesToTimber(reqBody, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
			return nil, err
		}

		startTime := time.Now()
		result, err = producerClient.Produce(ctx, &timber, grpcCallOption...)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, pAttr.producerAddr, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, profile.ProducerAddress, req, err)
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
	appName := req.Header.Get(AppNameHeaderName)
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
	newReq.Header.Add(RouterForwardingHeaderName, "1")
	fmt.Println("Forwarding to other router:", host+req.URL.Path)
	fmt.Println("Headers:", newReq.Header)

	// Send the request to the other router
	resp, err := p.client.Do(newReq)
	if err != nil {
		instrumentation.IncreaseForwardToOtherRouterFailed(appGroupName, appName, host)
	} else {
		instrumentation.IncreaseForwardToOtherRouterSuccess(appGroupName, appName, host)
	}
	return resp, err
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
