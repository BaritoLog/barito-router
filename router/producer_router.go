package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	pb "github.com/vwidjaya/barito-proto/producer"
)

const (
	AppSecretHeaderName      = "X-App-Secret"
	AppGroupSecretHeaderName = "X-App-Group-Secret"
	AppNameHeaderName        = "X-App-Name"
	KeyProducer              = "producer"
	AppNoProfilePath         = "api/producer_no_profile"
	AppNoSecretPath          = "api/no_secret"
)

type ProducerRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type producerRouter struct {
	addr                  string
	marketUrl             string
	profilePath           string
	profileByAppGroupPath string
	cacheBag              *cache.Cache
	client                *http.Client
	appCtx                *appcontext.AppContext
	producerStore         *ProducerStore
}

func NewProducerRouter(addr, marketUrl, profilePath string, profileByAppGroupPath string, appCtx *appcontext.AppContext) ProducerRouter {
	return &producerRouter{
		addr:                  addr,
		marketUrl:             marketUrl,
		profilePath:           profilePath,
		profileByAppGroupPath: profileByAppGroupPath,
		cacheBag:              cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:                createClient(),
		appCtx:                appCtx,
		producerStore:         NewProducerStore(),
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
	)
	span.SetTag("app-group", profile.ClusterName)

	//Collect all results and errors for all the produce endpoints.
	var produceResults []*pb.ProduceResult
	var produceErrors []error

	if req.Body != nil {
		reqBody, _ = ioutil.ReadAll(req.Body)
	}

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
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, req, err)
		return true
	}

	if profile == nil {
		onNoProfile(w)
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, req, err)
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
		logProduceError(instrumentation.ErrorConsulCall, profile.ClusterName, appGroupSecret, appName, req, err)
		return producerAttributes{}, err
	}
	if srv == nil {
		err = fmt.Errorf("Can't find service from consul: %s", KeyProducer)
		onConsulError(w, err)
		logProduceError(instrumentation.ErrorNoProducer, profile.ClusterName, appGroupSecret, appName, req, err)
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
		consulAddr:   "",
		producerAddr: profile.ProducerAddress,
		producerName: producerName,
		appSecret:    profile.AppSecret,
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

	// Check if the request has a "Content-Encoding" header with value "gzip"
	if req.Header.Get("Content-Encoding") == "gzip" {
		fmt.Println("+++++ New GZIP request", req.Header.Get("X-App-Name"))
		fmt.Println("prev reqBody", string(reqBody))
		// Decompress the gzip-encoded request body
		gzipReader, err := gzip.NewReader(bytes.NewReader(reqBody))
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorGzipDecompression, profile.ClusterName, appGroupSecret, appName, req, err)
			return nil, err
		}
		defer gzipReader.Close()

		// Read the decompressed request body into a buffer
		reqBody, err = ioutil.ReadAll(gzipReader)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorGzipDecompression, profile.ClusterName, appGroupSecret, appName, req, err)
			return nil, err
		}
		fmt.Println("after reqBody", string(reqBody))
	}

	if req.URL.Path == "/produce_batch" {
		timberCollection, err := ConvertBytesToTimberCollection(reqBody, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, req, err)
			return nil, err
		}

		startTime := time.Now()
		result, err = producerClient.ProduceBatch(ctx, &timberCollection)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, req, err)
			return nil, err
		}
		instrumentation.ObserveByteIngestion(profile.ClusterName, appName, reqBody)
		return result, nil

	}
	if req.URL.Path == "/produce" {
		timber, err := ConvertBytesToTimber(reqBody, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, req, err)
			return nil, err
		}

		startTime := time.Now()
		result, err = producerClient.Produce(ctx, &timber)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, req, err)
			return nil, err
		}
		instrumentation.ObserveByteIngestion(profile.ClusterName, appName, reqBody)
		return result, nil
	}

	return nil, fmt.Errorf("Invalid URL called - %s", req.URL.Path)
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
