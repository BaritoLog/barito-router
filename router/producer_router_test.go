package router

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/BaritoLog/go-boilerplate/httpkit"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/hashicorp/consul/api"
)

func TestProducerRouter_Ping(t *testing.T) {
	marketServer := NewTestServer(http.StatusOK, []byte(``))
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/ping", nil)

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
}

func TestProducerRouter_FetchError(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "some-secret")

	router := NewProducerRouter(":65500", "http://wrong-market", "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusBadGateway)
}

func TestProducerRouter_NoSecret(t *testing.T) {
	router := NewProducerRouter(":65500", "http://wrong-market", "profilePath")

	req, _ := http.NewRequest("GET", "/", nil)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusBadRequest)
}

func TestProducerRouter_NoProfile(t *testing.T) {
	marketServer := NewTestServer(http.StatusNotFound, []byte(``))
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "some-secret")

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestProducerRouter_WithAppGroupSecret_NoProfile(t *testing.T) {
	marketServer := NewTestServer(http.StatusNotFound, []byte(``))
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("App-Name", "some-name")

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestProducerRouter_ConsulError(t *testing.T) {
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHost: "wrong-consul",
	})
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "some-secret")

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusFailedDependency)
}

func TestProducerRouter_WithAppSecret(t *testing.T) {

	targetServer := NewTestServer(http.StatusTeapot, []byte("some-target"))
	defer targetServer.Close()
	host, port := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		api.CatalogService{
			ServiceAddress: host,
			ServicePort:    port,
		},
	})
	defer consulServer.Close()

	host, port = httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHost: fmt.Sprintf("%s:%d", host, port),
	})
	defer marketServer.Close()

	testPayload := []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)
	FatalIfWrongResponseStatus(t, resp, http.StatusTeapot)
	FatalIfWrongResponseBody(t, resp, "some-target")
}

func TestProducerRouter_WithAppGroupSecret(t *testing.T) {

	targetServer := NewTestServer(http.StatusTeapot, []byte("some-target"))
	defer targetServer.Close()
	host, port := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		api.CatalogService{
			ServiceAddress: host,
			ServicePort:    port,
		},
	})
	defer consulServer.Close()

	host, port = httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHost: fmt.Sprintf("%s:%d", host, port),
	})
	defer marketServer.Close()

	testPayload := []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("App-Name", "some-name")

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusTeapot)
	FatalIfWrongResponseBody(t, resp, "some-target")
}
