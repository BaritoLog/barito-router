package router

import (
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
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

func onRpcError(w http.ResponseWriter, err error) string {
	httpCode := http.StatusBadGateway

	st, ok := status.FromError(err)
	if ok {
		httpCode = runtime.HTTPStatusFromCode(st.Code())
	}

	w.WriteHeader(httpCode)
	w.Write([]byte(st.Message()))
	return st.Message()
}

func onRpcSuccess(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func logProduceError(context, clusterName, appGroupSecret, appName string, err error) {
	maskedAppGroupSecret := appGroupSecret[0:6]
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	msg := fmt.Sprintf("Got error clusterName=%q, appgroupSecret=%q, appname=%q, context=%q, error=%q", clusterName, maskedAppGroupSecret, appName, context, errorMsg)
	log.Errorf("%s", msg)
}
