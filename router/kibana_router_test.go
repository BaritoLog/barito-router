package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/go-boilerplate/httpkit"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/gorilla/mux"
	newrelic "github.com/newrelic/go-agent"
	"golang.org/x/time/rate"
)

func TestKibanaRouter_Ping(t *testing.T) {
	marketServer := NewTestServer(http.StatusOK, []byte(``))
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", appCtx)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/ping", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
}

func TestKibanaRouter_FetchError(t *testing.T) {
	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":65500", "http://wrong-market", "abc", "profilePath", "authorizePath", appCtx)

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

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", appCtx)
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", strings.NewReader(""))
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestKibanaRouter_K8s(t *testing.T) {
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

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", appCtx)

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
func TestRateLimiter(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limitedHandler := RateLimiter(limiter)(handler)

	// Test when the limiter allows the request
	req1, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	resp1 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(resp1, req1)

	if resp1.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp1.Code)
	}

	// Test when the limiter blocks the request
	req2, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	resp2 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(resp2, req2)

	if resp2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status code %d, but got %d", http.StatusTooManyRequests, resp2.Code)
	}
}

func TestKibanaRouter_ServeElasticsearch(t *testing.T) {
	targetServer := NewTestServer(http.StatusTeapot, []byte("elasticsearch response"))
	defer targetServer.Close()
	host, port := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ElasticsearchAddress: fmt.Sprintf("%s:%d", host, port),
		ElasticsearchStatus:  "ACTIVE",
		AppGroupSecret:       "mock-secret",
	})
	defer marketServer.Close()

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewKibanaRouter(":45500", marketServer.URL, "abc", "profilePath", "authorizePath", appCtx)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/elasticsearch/my_cluster/_search", strings.NewReader(""))
	req.Header.Set("App-Group-Secret", "mock-secret")

	r := mux.NewRouter()
	r.HandleFunc("/elasticsearch/{cluster_name}/{es_endpoint:.*}", router.ServeElasticsearch)
	resp := RecordResponse(r.ServeHTTP, req)

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != http.StatusTeapot && resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status code %d or %d, but got %d. Response body: %s", http.StatusTeapot, http.StatusInternalServerError, resp.StatusCode, bodyStr)
	}

	expectedBody := "elasticsearch response"
	if bodyStr != expectedBody && bodyStr != "Elasticsearch is unreachable\n" {
		t.Fatalf("Expected response body %q or 'Elasticsearch is unreachable', but got %q", expectedBody, bodyStr)
	}
}
