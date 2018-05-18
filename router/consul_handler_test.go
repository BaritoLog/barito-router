package router

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestConsulHandler_New(t *testing.T) {
	consul := NewConsulHandler()

	FatalIf(t, consul == nil, "Consul can't be nil")
}

func TestConsulHandler_Service_InvalidConsulAddress(t *testing.T) {
	consul := NewConsulHandler()
	_, err := consul.Service("invalid-consul-address", "some-service")

	FatalIf(t, !strings.Contains(err.Error(), "no such host"), "wrong error")
}

func TestConsulHandler_Service_NoService(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte(`[]`))

	consul := NewConsulHandler()
	_, err := consul.Service(ts.URL, "some-service")
	FatalIfWrongError(t, err, "No consul service found for 'some-service'")
}

func TestConsulHandler_Service(t *testing.T) {

	ts := NewHttpTestServer(http.StatusOK, []byte(`[
  {
    "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
    "Node": "foobar",
    "Address": "192.168.10.10",
    "Datacenter": "dc1",
    "TaggedAddresses": {
      "lan": "192.168.10.10",
      "wan": "10.0.10.10"
    },
    "NodeMeta": {
      "somekey": "somevalue"
    },
    "CreateIndex": 51,
    "ModifyIndex": 51,
    "ServiceAddress": "172.17.0.3",
    "ServiceEnableTagOverride": false,
    "ServiceID": "32a2a47f7992:nodea:5000",
    "ServiceName": "foobar",
    "ServicePort": 5000,
    "ServiceMeta": {
        "foobar_meta_value": "baz"
    },
    "ServiceTags": [
      "tacos"
    ]
  }
]`))

	consul := NewConsulHandler()
	service, err := consul.Service(ts.URL, "some-service")
	FatalIfError(t, err)
	FatalIf(t, service.Address != "192.168.10.10", "wrong service.Address")
	FatalIf(t, service.ID != "40e4a748-2192-161a-0510-9bf59fe950b5", "wrong service.ID")

}
