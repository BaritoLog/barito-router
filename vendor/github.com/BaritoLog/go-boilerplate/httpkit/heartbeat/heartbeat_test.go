package heartbeat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/imantung/go-boilerplate/testkit"
)

func TestHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/heartbeat", nil)
	rec := httptest.NewRecorder()

	http.HandlerFunc(Handler).ServeHTTP(rec, req)
	FatalIfWrongHttpCode(t, rec, http.StatusOK)

	var message HeartbeatMessage
	json.Unmarshal(rec.Body.Bytes(), &message)
	FatalIf(t, message.Time.IsZero(), "message.Time is nil")
}
