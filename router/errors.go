package router

import "net/http"

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
	w.WriteHeader(http.StatusFailedDependency)
	w.Write([]byte(err.Error()))
}
