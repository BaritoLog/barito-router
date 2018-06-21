package testkit

import (
	"fmt"
	"net/http/httptest"
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

func FatalIfWrongHttpCode(t T, rec *httptest.ResponseRecorder, code int) {
	if rec.Code != code {
		message := fmt.Sprintf("wrong http code: %d", rec.Code)
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
