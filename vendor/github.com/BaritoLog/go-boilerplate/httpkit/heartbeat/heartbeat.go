package heartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HeartbeatMessage struct {
	Time time.Time `json:"time"`
}

func Handler(respWriter http.ResponseWriter, req *http.Request) {
	b, _ := json.Marshal(HeartbeatMessage{
		Time: time.Now(),
	})

	fmt.Fprintf(respWriter, "%s", string(b))
	respWriter.WriteHeader(http.StatusOK)
}
