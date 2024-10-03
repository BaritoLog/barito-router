package router

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type AuditLog struct {
	Type            string `json:"type"`
	Timestamp       string `json:"timestamp"`
	RequestHost     string `json:"request_host"`
	RequestMethod   string `json:"request_method"`
	RequestPath     string `json:"request_path"`
	AppSecret       string `json:"app_secret"`
	ClusterName     string `json:"cluster_name"`
	Status          int    `json:"status"`
	RemoteIP        string `json:"remote_ip"`
	UserAgent       string `json:"user_agent"`
	RequestDuration string `json:"request_duration"`
	ResponseSize    int    `json:"response_size"`
}

func LogAudit(req *http.Request, esRes *http.Response, body []byte, appSecret, clusterName string, duration time.Duration) {
	auditLog := AuditLog{
		Type:            "audit",
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		RequestHost:     req.Host,
		RequestMethod:   req.Method,
		RequestPath:     req.URL.Path,
		AppSecret:       maskAppSecret(appSecret),
		ClusterName:     clusterName,
		Status:          esRes.StatusCode,
		RemoteIP:        req.RemoteAddr,
		UserAgent:       req.UserAgent(),
		RequestDuration: duration.String(),
		ResponseSize:    len(body),
	}
	auditLogJSON, err := json.Marshal(auditLog)
	if err != nil {
		log.Printf("Failed to marshal audit log: %v", err)
		return
	}
	log.Println(string(auditLogJSON))
}

func maskAppSecret(appSecret string) string {
	if len(appSecret) > 4 {
		return appSecret[:4] + "*****"
	}
	return appSecret
}
