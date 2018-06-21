package testkit

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestHttpServer(t *testing.T) {

	ts := NewTestServer(http.StatusTeapot, []byte("some-result"))

	res, err := http.Get(ts.URL)
	FatalIfError(t, err)

	b, _ := ioutil.ReadAll(res.Body)
	FatalIf(t, res.StatusCode != http.StatusTeapot, "wrong res.StatusCode")
	FatalIf(t, string(b) != "some-result", "wrong.res.Body")

}

func TestRecord(t *testing.T) {

	var got *http.Request
	req, _ := http.NewRequest(http.MethodPost, "http://some-url", strings.NewReader("some-req-body"))
	handler := func(rw http.ResponseWriter, req *http.Request) {
		got = req
		rw.WriteHeader(http.StatusTeapot)
		rw.Write([]byte("hello mama"))
	}

	rr := RecordResponse(handler, req)
	b, _ := ioutil.ReadAll(got.Body)

	FatalIf(t, rr.Code != http.StatusTeapot, "wrong rr.Code")
	FatalIf(t, string(rr.Body.Bytes()) != "hello mama", "wrong rr.Body")
	FatalIf(t, got.Method != http.MethodPost, "wrong got.Method")
	FatalIf(t, got.URL.String() != "http://some-url", "wrong got.URL")
	FatalIf(t, string(b) != "some-req-body", "wrong got.Body")
}
