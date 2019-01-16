package instrumentation

import (
	"net/http"
	"fmt"

	"github.com/newrelic/go-agent"
)

func RunTransaction(app newrelic.Application, path string, w http.ResponseWriter, req *http.Request) {
	if app != nil {
		txn := app.StartTransaction(fmt.Sprintf("/%s", path), w, req)
		defer txn.End()
	}
}
