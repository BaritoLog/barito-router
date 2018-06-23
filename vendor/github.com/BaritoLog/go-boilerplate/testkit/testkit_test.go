package testkit

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestFatalIfError(t *testing.T) {
	dt := &dummyT{}

	FatalIfError(dt, fmt.Errorf("some-error"))

	if dt.str != "testkit_test.go:14: some-error\n" || !dt.isFail {
		t.Fatalf("FatalIfError got error: %s", dt.str)
	}
}

func TestFatalIfWrongError(t *testing.T) {
	dt := &dummyT{}
	FatalIfWrongError(dt, fmt.Errorf("some-error"), "not-some-error")

	if dt.str != "testkit_test.go:23: Wrong error message: some-error\n" || !dt.isFail {
		t.Fatalf("FatalIfWrongError got message: %s", dt.str)
	}
}

func TestFatalIfWrongError_ErrorIsNil(t *testing.T) {
	dt := &dummyT{}
	FatalIfWrongError(dt, nil, "not-some-error")

	if dt.str != "testkit_test.go:32: no expected error\n" || !dt.isFail {
		t.Fatalf("FatalIfWrongError got message: %s", dt.str)
	}
}

func TestFatalIf(t *testing.T) {
	dt := &dummyT{}
	FatalIf(dt, true, "some-error: %s", "message")

	if dt.str != "testkit_test.go:41: some-error: message\n" || !dt.isFail {
		t.Fatalf("FatalIf got message: %s", dt.str)
	}
}

func TestFatalIf_FalseCondition(t *testing.T) {
	dt := &dummyT{}
	FatalIf(dt, false, "some-error: %s", "message")

	if dt.str != "" || dt.isFail {
		t.Fatalf("FatalIf should not fail or return any message")
	}
}

func TestFatalIfWrongResponseStatus(t *testing.T) {
	dt := &dummyT{}

	FatalIfWrongResponseStatus(dt,
		&http.Response{
			StatusCode: http.StatusTeapot,
		},
		http.StatusOK)

	if dt.str != "testkit_test.go:60: wrong response status code: 418\n" || !dt.isFail {
		t.Fatalf("FatalIfWrongResponseStatus got message: %s", dt.str)
	}
}

func TestFatalIfWrongResponseBody(t *testing.T) {
	dt := &dummyT{}

	FatalIfWrongResponseBody(dt,
		&http.Response{
			Body: ioutil.NopCloser(strings.NewReader("wrong-body")),
		},
		"expected-body")

	if dt.str != "testkit_test.go:74: wrong response body: wrong-body\n" || !dt.isFail {
		t.Fatalf("FatalIfWrongResponseStatus got message: %s", dt.str)
	}

}
