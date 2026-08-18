package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	pb "github.com/bentol/barito-proto/producer"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"
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

	ErrorDoubleRouterForward = "Request already forwarded from another router, skipping forwarding again."

	// VictoriaLogsCommandName is the hystrix circuit-breaker command name for
	// the fire-and-forget double-write to VictoriaLogs.
	VictoriaLogsCommandName = "victorialogs"
	// K8sMetadataMarker is the token that identifies a k8s log (skipped for VictoriaLogs).
	K8sMetadataMarker = `"k8s_metadata"`
	// VictoriaLogsTimeout bounds a single fire-and-forget write to VictoriaLogs.
	VictoriaLogsTimeout = 10 * time.Second
)

// vlogsJob is one enqueued double-write. Body is held compressed (as received)
// so the queue stays small and decompression happens off the produce path.
type vlogsJob struct {
	path    string
	body    []byte
	isGzip  bool
	appName string // X-App-Name header; exact app on the app-group path
	profile *Profile
}

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
	vlogsClient                       heimdall.Doer
	vlogsJobs                         chan vlogsJob
}

// newVictoriaLogsClient builds the circuit-broken HTTP client for the
// fire-and-forget double-write to VictoriaLogs. When VictoriaLogs is slow or
// down, the breaker opens and writes fail fast instead of piling up goroutines.
func newVictoriaLogsClient() heimdall.Doer {
	return hystrix.NewClient(
		hystrix.WithHTTPTimeout(VictoriaLogsTimeout),
		hystrix.WithHystrixTimeout(VictoriaLogsTimeout),
		hystrix.WithHTTPClient(createClient()),
		hystrix.WithCommandName(VictoriaLogsCommandName),
		hystrix.WithMaxConcurrentRequests(100),
		hystrix.WithRequestVolumeThreshold(20),
		hystrix.WithErrorPercentThreshold(25),
		hystrix.WithSleepWindow(5000),
	)
}

func NewProducerRouter(addr, marketUrl, profilePath string, profileByAppGroupPath string, appCtx *appcontext.AppContext) ProducerRouter {
	p := &producerRouter{
		addr:                              addr,
		marketUrl:                         marketUrl,
		profilePath:                       profilePath,
		profileByAppGroupPath:             profileByAppGroupPath,
		cacheBag:                          cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:                            createClient(),
		appCtx:                            appCtx,
		producerStore:                     NewProducerStore(),
		isRouterLocationForwardingEnabled: len(config.RouterLocationForwardingMap) > 0,
		vlogsClient:                       newVictoriaLogsClient(),
	}
	if config.VictoriaLogsUrl != "" {
		p.startVictoriaLogsWorkers()
	}
	return p
}

// startVictoriaLogsWorkers launches the fixed worker pool that drains the
// double-write queue. Workers live for the process lifetime.
func (p *producerRouter) startVictoriaLogsWorkers() {
	p.vlogsJobs = make(chan vlogsJob, config.VictoriaLogsQueueSize)
	for i := 0; i < config.VictoriaLogsWorkers; i++ {
		go p.vlogsWorker()
	}
}

// enqueueVictoriaLogs hands a job to the worker pool without ever blocking:
// a full queue drops the log and counts it, so the produce path is unaffected.
func (p *producerRouter) enqueueVictoriaLogs(job vlogsJob) {
	select {
	case p.vlogsJobs <- job:
	default:
		instrumentation.IncreaseVictoriaLogsDropped(job.profile.ClusterName)
	}
}

func (p *producerRouter) vlogsWorker() {
	for job := range p.vlogsJobs {
		body := job.body
		if job.isGzip {
			body = decompressGzip(body)
		}
		p.handleProduceToVictoriaLogs(context.Background(), job.path, body, job.appName, job.profile)
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
		// check if the router is eligible for forwarding, based on producer_location & this router env variables
		if host, isEligible := p.isEligibleForRouterLocationForwarding(profile); isEligible {
			// make sure the request is not forwarded from another router
			if req.Header.Get(RouterForwardingHeaderName) != "" {
				p.onDoubleRouterForward(w, req, profile.ClusterName)
				return
			}

			p.onEligibleForwarding(w, req, host, profile.ClusterName, reqBody)
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

	if config.VictoriaLogsUrl != "" {
		// Fire-and-forget: hand the write to the worker pool and move on. A
		// non-blocking send means a full queue drops the log instead of ever
		// blocking the produce critical path. Body stays compressed here; the
		// worker decompresses off the hot path.
		p.enqueueVictoriaLogs(vlogsJob{
			path:    req.URL.Path,
			body:    reqBody,
			isGzip:  req.Header.Get("Content-Encoding") == "gzip",
			appName: req.Header.Get(AppNameHeaderName),
			profile: profile,
		})
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

func isK8sLog(reqBody []byte) bool {
	return bytes.Contains(reqBody, []byte(K8sMetadataMarker))
}

func buildVictoriaLogsBody(path string, reqBody []byte) []byte {
	if path == "/produce_batch" {
		var batch map[string][]json.RawMessage
		if err := json.Unmarshal(reqBody, &batch); err != nil {
			return nil
		}

		items := batch["items"]
		if len(items) == 0 {
			return nil
		}

		var buf bytes.Buffer
		for _, item := range items {
			if isK8sLog(item) {
				continue
			}
			if err := json.Compact(&buf, item); err != nil {
				continue
			}
			buf.WriteByte('\n')
		}

		if buf.Len() == 0 {
			return nil
		}
		return buf.Bytes()
	}

	if isK8sLog(reqBody) {
		return nil
	}
	result := make([]byte, len(reqBody)+1)
	copy(result, reqBody)
	result[len(reqBody)] = '\n'
	return result
}

// decompressGzip inflates a gzip body, returning the original bytes if it is
// not valid gzip (best-effort; the double-write must not fail loudly).
func decompressGzip(b []byte) []byte {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return b
	}
	defer r.Close()
	if d, err := io.ReadAll(r); err == nil {
		return d
	}
	return b
}

func (p *producerRouter) handleProduceToVictoriaLogs(reqCtx context.Context, path string, reqBody []byte, appName string, profile *Profile) {
	// Runs in its own goroutine (fire-and-forget). An unrecovered panic here
	// would crash the whole router, so contain it — the double-write is best-effort.
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("recovered from panic in VictoriaLogs double-write: %v", r)
			if profile != nil {
				instrumentation.IncreaseVictoriaLogsFailed(profile.ClusterName)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(reqCtx, VictoriaLogsTimeout)
	defer cancel()

	body := buildVictoriaLogsBody(path, reqBody)
	if body == nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.VictoriaLogsUrl, bytes.NewReader(body))
	if err != nil {
		log.Errorf("failed to create VictoriaLogs request: %v", err)
		instrumentation.IncreaseVictoriaLogsFailed(profile.ClusterName)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// Tag every ingested log with the origin the router knows (from the profile,
	// not the dynamic client payload) so VictoriaLogs partitions streams by
	// cluster/app instead of collapsing everything into the default stream.
	if appName == "" {
		appName = profile.Name
	}
	req.Header.Set("VL-Stream-Fields", "cluster_name,application_name,app_group")
	req.Header.Set("VL-Extra-Fields", fmt.Sprintf("cluster_name=%s,application_name=%s,app_group=%s",
		profile.ClusterName, appName, profile.AppGroup))

	resp, err := p.vlogsClient.Do(req)
	if err != nil {
		log.Errorf("failed to forward to VictoriaLogs: %v", err)
		instrumentation.IncreaseVictoriaLogsFailed(profile.ClusterName)
		return
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		log.Errorf("VictoriaLogs returned status %d for cluster %s", resp.StatusCode, profile.ClusterName)
		instrumentation.IncreaseVictoriaLogsFailed(profile.ClusterName)
		return
	}
	instrumentation.IncreaseVictoriaLogsSuccess(profile.ClusterName)
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

func (p *producerRouter) onDoubleRouterForward(w http.ResponseWriter, req *http.Request, clusterName string) {
	slog.Error(ErrorDoubleRouterForward)
	instrumentation.IncreaseDoubleRouterForward(clusterName, req.Header.Get(AppNameHeaderName))
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(ErrorDoubleRouterForward))
}

func (p *producerRouter) onEligibleForwarding(w http.ResponseWriter, req *http.Request, otherRouterHost, clusterName string, reqBody []byte) {
	appName := req.Header.Get(AppNameHeaderName)
	resp, err := p.forwardToOtherRouter(otherRouterHost, req, reqBody, clusterName)
	if err != nil {
		slog.Error("Error forwarding to other router", slog.String("error", err.Error()))
		instrumentation.IncreaseForwardToOtherRouterFailed(clusterName, appName, otherRouterHost)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response from other router:", resp.StatusCode)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Error("Error copying response body", slog.String("error", err.Error()))
		instrumentation.IncreaseForwardToOtherRouterFailed(clusterName, appName, otherRouterHost)
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
