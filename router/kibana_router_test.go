package router

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/BaritoLog/go-boilerplate/httpkit"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/hashicorp/consul/api"
)

func TestKibanaRouter_Ping(t *testing.T) {
	marketServer := NewTestServer(http.StatusOK, []byte(``))
	defer marketServer.Close()

	router := NewKibanaRouter(":45500", marketServer.URL, "profilePath", "authorizePath", "")
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/ping", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
}

func TestKibanaRouter_FetchError(t *testing.T) {
	router := NewKibanaRouter(":65500", "http://wrong-market", "profilePath", "authorizePath", "")

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusBadGateway)
}

func TestKibanaRouter_NoProfile(t *testing.T) {
	marketServer := NewTestServer(http.StatusNotFound, []byte(``))
	defer marketServer.Close()

	router := NewKibanaRouter(":45500", marketServer.URL, "profilePath", "authorizePath", "")
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestKibanaRouter_ConsulError(t *testing.T) {
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHost: "wrong-consul",
	})
	defer marketServer.Close()

	router := NewKibanaRouter(":45500", marketServer.URL, "profilePath", "authorizePath", "")
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIf(t, resp.StatusCode != http.StatusFailedDependency, "Wrong response status code")
}

func TestKibanaRouter(t *testing.T) {

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

	router := NewKibanaRouter(":45500", marketServer.URL, "profilePath", "authorizePath", "")

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))

	resp := RecordResponse(router.ServeHTTP, req)
	FatalIfWrongResponseStatus(t, resp, http.StatusTeapot)
	FatalIfWrongResponseBody(t, resp, "some-target")
}

func TestGetClustername(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/path", strings.NewReader(""))

	cluster_name := KibanaGetClustername(req)
	FatalIf(t, cluster_name != "path", "%s != %s", cluster_name, "path")
}
