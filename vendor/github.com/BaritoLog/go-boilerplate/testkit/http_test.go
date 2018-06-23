package testkit

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestNewTestServer(t *testing.T) {
	ts := NewTestServer(http.StatusTeapot, []byte("some-result"))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	FatalIfError(t, err)
	FatalIfWrongResponseStatus(t, res, http.StatusTeapot)
	FatalIfWrongResponseBody(t, res, "some-result")
}

func TestNewJsonTestServer(t *testing.T) {
	person := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{"iman", 31}

	ts := NewJsonTestServer(http.StatusTeapot, person)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	FatalIfError(t, err)
	FatalIfWrongResponseBody(t, res, `{"name":"iman","age":31}`)
}

func TestRecordResponse(t *testing.T) {

	var got *http.Request
	req, _ := http.NewRequest(http.MethodPost, "http://some-url", strings.NewReader("some-req-body"))
	handler := func(rw http.ResponseWriter, req *http.Request) {
		got = req
		rw.WriteHeader(http.StatusTeapot)
		rw.Write([]byte("hello mama"))
	}

	resp := RecordResponse(handler, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusTeapot)
	FatalIfWrongResponseBody(t, resp, "hello mama")

	b, _ := ioutil.ReadAll(got.Body)
	FatalIf(t, string(b) != "some-req-body", "wrong got.Body")
	FatalIf(t, got.Method != http.MethodPost, "wrong got.Method")
	FatalIf(t, got.URL.String() != "http://some-url", "wrong got.URL")

}
