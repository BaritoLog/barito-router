package router

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"log/slog"
)

type AuditLog struct {
	Type            string `json:"type"`
	Timestamp       string `json:"timestamp"`
	RequestHost     string `json:"request_host"`
	RequestMethod   string `json:"request_method"`
	RequestPath     string `json:"request_path"`
	ClusterName     string `json:"cluster_name"`
	Status          int    `json:"status"`
	RemoteIP        string `json:"remote_ip"`
	UserAgent       string `json:"user_agent"`
	RequestDuration string `json:"request_duration"`
	ResponseSize    int    `json:"response_size"`
}

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func LogAudit(req *http.Request, esRes *http.Response, body []byte, appSecret, clusterName string, duration time.Duration) {
	auditLog := AuditLog{
		Type:            "audit",
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		RequestHost:     req.Host,
		RequestMethod:   req.Method,
		RequestPath:     req.URL.RequestURI(),
		ClusterName:     clusterName,
		Status:          esRes.StatusCode,
		RemoteIP:        req.RemoteAddr,
		UserAgent:       req.UserAgent(),
		RequestDuration: duration.String(),
		ResponseSize:    len(body),
	}
	auditLogJSON, err := json.Marshal(auditLog)
	if err != nil {
		logger.Error("Failed to marshal audit log", slog.String("error", err.Error()))
		return
	}
	logger.Info(string(auditLogJSON))
}
