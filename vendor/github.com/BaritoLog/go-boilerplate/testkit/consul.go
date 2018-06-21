package testkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/consul/api"
)

func NewConsulCatalogTestServer(services []*api.CatalogService) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := json.Marshal(services)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
}
