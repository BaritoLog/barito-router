package router

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestProducerProxyHandler(t *testing.T) {

	url := &url.URL{
		Scheme: "http",
		Host:   "localhost:12345",
	}

	profile := Profile{}
	proxyHandler := NewProducerProxyHandler(url, profile)

	body := strings.NewReader(`
	{
		"hello":"world"
	}
	`)
	req, _ := http.NewRequest("GET", "http://example.com", body)
	proxyHandler.Director(req)

	var timber map[string]interface{}
	b, _ := ioutil.ReadAll(req.Body)

	err := json.Unmarshal(b, &timber)
	if err != nil {
		fmt.Println(err)
		return
	}

	b, _ = json.Marshal(timber["_ctx"])

	refCtx := map[string]interface{}{
		"es_index_prefix":          "",
		"app_max_tps":              0,
		"kafka_topic":              "",
		"kafka_partition":          0,
		"kafka_replication_factor": 0,
		"es_document_type":         "",
	}
	bref, _ := json.Marshal(refCtx)
	FatalIf(t, string(b) != string(bref), "Context not found or invalid context inserted")
}
