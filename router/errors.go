package router

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/BaritoLog/barito-router/instrumentation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/status"
)

// TODO: rename to onFetchError
func onTradeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(err.Error()))
}

func onNoProfile(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("No Profile"))
}

func onNoSecret(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("No Secret"))
}

func onConsulError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusFailedDependency)
	w.Write([]byte(err.Error()))
}

func onKafkaPixyError(w http.ResponseWriter, err error) {
	// TODO: change status code
	w.WriteHeader(http.StatusFailedDependency)
	w.Write([]byte(err.Error()))
}

func onAuthorizeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func onRpcError(w http.ResponseWriter, errors []error) string {
	httpCode := http.StatusBadGateway

	st, ok := status.FromError(errors[0])
	if ok {
		httpCode = runtime.HTTPStatusFromCode(st.Code())
	}
	w.WriteHeader(httpCode)

	var responseMsg bytes.Buffer
	for _, err := range errors {
		responseMsg.WriteString(err.Error() + "\n")
	}
	w.Write(responseMsg.Bytes())
	return responseMsg.String()
}

func onRpcSuccess(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func logProduceError(context, clusterName, appGroupSecret, appName, producerAddr string, r *http.Request, err error, span trace.Span) {
	maskedAppGroupSecret := appGroupSecret
	if len(maskedAppGroupSecret) > 6 {
		maskedAppGroupSecret = appGroupSecret[0:6]
	}
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	msg := fmt.Sprintf("Got error clusterName=%q, appgroupSecret=%q, appname=%q, context=%q, error=%q", clusterName, maskedAppGroupSecret, appName, context, errorMsg)
	log.Errorf("%s", msg)

	span.SetStatus(codes.Error, errorMsg)

	instrumentation.IncreaseProducerRequestError(clusterName, appName, producerAddr, r, context)
}
