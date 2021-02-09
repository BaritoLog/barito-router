package router

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/go-boilerplate/httpkit"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/hashicorp/consul/api"
	newrelic "github.com/newrelic/go-agent"
)

func TestKibanaRouter_Ping(t *testing.T) {
	marketServer := NewTestServer(http.StatusOK, []byte(``))
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", "", appCtx)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/ping", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
}

func TestKibanaRouter_FetchError(t *testing.T) {
	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":65500", "http://wrong-market", "abc", "profilePath", "authorizePath", "", appCtx)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusBadGateway)
}

func TestKibanaRouter_NoProfile(t *testing.T) {
	marketServer := NewTestServer(http.StatusNotFound, []byte(``))
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", "", appCtx)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestKibanaRouter_ConsulError(t *testing.T) {
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts: []string{"wrong-consul"},
	})
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", "", appCtx)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIf(t, resp.StatusCode != http.StatusFailedDependency, "Wrong response status code")
}

func TestKibanaRouter(t *testing.T) {

	targetServer := NewTestServer(http.StatusTeapot, []byte("some-target"))
	defer targetServer.Close()
	host, port := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		{
			ServiceAddress: host,
			ServicePort:    port,
		},
	})
	defer consulServer.Close()

	host, port = httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts: []string{fmt.Sprintf("%s:%d", host, port)},
	})
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", "", appCtx)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))

	resp := RecordResponse(router.ServeHTTP, req)
	FatalIfWrongResponseStatus(t, resp, http.StatusTeapot)
	FatalIfWrongResponseBody(t, resp, "some-target")
}

func TestKibanaRouter_k8s(t *testing.T) {
	targetServer := NewTestServer(http.StatusTeapot, []byte("some-target"))
	defer targetServer.Close()
	host, port := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		KibanaAddress: fmt.Sprintf("%s:%d", host, port),
	})
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", "", appCtx)

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
