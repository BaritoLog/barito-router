package router

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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
