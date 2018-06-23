package testkit

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
)

func FatalIfError(t T, err error) {
	if err != nil {
		fatal(t, err.Error(), 1)
	}
}

func FatalIfWrongError(t T, err error, message string) {
	if err == nil {
		fatal(t, "no expected error", 1)
		return
	}

	if !strings.Contains(err.Error(), message) {
		fatal(
			t,
			fmt.Sprintf("Wrong error message: %s", err.Error()),
			1,
		)
	}
}

func FatalIfWrongResponseStatus(t T, resp *http.Response, statusCode int) {
	if resp.StatusCode != statusCode {
		message := fmt.Sprintf("wrong response status code: %d", resp.StatusCode)
		fatal(t, message, 1)
	}
}

func FatalIfWrongResponseBody(t T, resp *http.Response, body string) {
	b, _ := ioutil.ReadAll(resp.Body)
	s := string(b)
	if s != body {
		message := fmt.Sprintf("wrong response body: %s", s)
		fatal(t, message, 1)
	}
}

func FatalIf(t T, condition bool, format string, v ...interface{}) {
	if condition {
		message := fmt.Sprintf(format, v...)
		fatal(t, message, 1)
	}
}

func fatal(t T, message string, funcLevel int) {
	_, file, no, ok := runtime.Caller(funcLevel + 1)
	if ok {
		simpleFileName := file[strings.LastIndex(file, "/")+1:]
		message = fmt.Sprintf("%s:%d: %s", simpleFileName, no, message)
	}

	t.Logf("%s\n", message)
	t.FailNow()
}
