package config

import (
	"strings"
	"time"

	"github.com/BaritoLog/go-boilerplate/envkit"
)

const (
	Name    = "Barito Router"
	Version = "0.7.1"

	EnvProducerRouterAddress             = "BARITO_PRODUCER_ROUTER"
	EnvKibanaRouterAddress               = "BARITO_KIBANA_ROUTER"
	EnvProducerPort                      = "BARITO_PRODUCER_PORT"
	EnvBaritoMarketUrl                   = "BARITO_MARKET_URL"
	EnvBaritoViewerUrl                   = "BARITO_VIEWER_URL"
	EnvBaritoMarketAccessToken           = "BARITO_MARKET_ACCESS_TOKEN"
	EnvBaritoProfileApiPath              = "BARITO_PROFILE_API_PATH"
	EnvBaritoProfileApiByAppGroupPath    = "BARITO_PROFILE_API_BY_APP_GROUP_PATH"
	EnvBaritoAuthorizeApiPath            = "BARITO_AUTHORIZE_API_PATH"
	EnvBaritoProfileApiByClusternamePath = "BARITO_PROFILE_API_BY_CLUSTERNAME_PATH"
	EnvNewRelicAppName                   = "BARITO_NEW_RELIC_APP_NAME"
	EnvNewRelicLicenseKey                = "BARITO_NEW_RELIC_LICENSE_KEY"
	EnvNewRelicEnabled                   = "BARITO_NEW_RELIC_ENABLED"
	EnvCacheExpirationTimeSeconds        = "BARITO_CACHE_EXPIRATION_TIME_IN_SECONDS"
	EnvBackupCacheExpirationTimeHours    = "BARITO_BACKUP_CACHE_EXPIRATION_TIME_IN_HOURS"
	EnvEnableTracing                     = "BARITO_ENABLE_TRACING"
	EnvEnableSSO                         = "BARITO_ENABLE_SSO"
	EnvSSORedirectPath                   = "BARITO_SSO_REDIRECT_PATH"
	EnvSSOClientID                       = "BARITO_SSO_CLIENT_ID"
	EnvSSOClientSecret                   = "BARITO_SSO_CLIENT_SECRET"
	EnvHMACJWTSecretString               = "BARITO_HMAC_JWT_SECRET_STRING"
	EnvAllowedDomains                    = "BARITO_ALLOWED_DOMAINS"
	EnvJaegerServiceName                 = "BARITO_JAEGER_SERVICE_NAME"
	EnvAllowedEndpoints                  = "BARITO_ALLOWED_ENDPOINTS"

	DefaultProducerRouterAddress             = ":8081"
	DefaultKibanaRouterAddress               = ":8083"
	DefaultProducerPort                      = ""
	DefaultBaritoMarketUrl                   = "http://localhost:8080"
	DefaultBaritoViewerUrl                   = "http://localhost:8083"
	DefaultBaritoMarketAccessToken           = ""
	DefaultBaritoProfileApiPath              = "api/profile"
	DefaultBaritoProfileApiByAppGroupPath    = "api/profile_by_app_group"
	DefaultBaritoAuthorizeApiPath            = "api/authorize"
	DefaultBaritoProfileApiByClusternamePath = "api/v2/profile_by_cluster_name"
	DefaultJaegerServiceName                 = "barito_router"
	DefaultNewRelicAppName                   = "barito_router"
	DefaultNewRelicLicenseKey                = ""
	DefaultEnableTracing                     = false
	DefaultEnableSSO                         = true
	DefaultSSORedirectPath                   = "/auth/callback"
	DefaultSSOClientID                       = ""
	DefaultSSOClientSecret                   = ""
	DefaultHMACJWTSecretString               = ""
	DefaultAllowedDomains                    = ""
	DefaultNewRelicEnabled                   = false
	DefaultCacheExpirationTimeSeconds        = 60
	DefaultBackupCacheExpirationTimeHours    = 48
	DefaultAllowedEndpoints                  = "_search,*/_search,_search/scroll,_doc,_cat/indices,_cat/health,_eql/search,_mget,_index,_ingest/pipeline"
)

var (
	RouterAddress                  string
	ProducerPort                   string
	KibanaRouterAddress            string
	BaritoMarketUrl                string
	BaritoViewerUrl                string
	BaritoMarketAccessToken        string
	ProfileApiPath                 string
	ProfileApiByAppGroupPath       string
	AuthorizeApiPath               string
	ProfileApiByClusternamePath    string
	JaegerServiceName              string
	NewRelicAppName                string
	NewRelicLicenseKey             string
	NewRelicEnabled                bool
	EnableTracing                  bool
	EnableSSO                      bool
	SSORedirectPath                string
	SSOClientID                    string
	SSOClientSecret                string
	HMACJWTSecretString            string
	AllowedDomains                 string
	CacheExpirationTimeSeconds     time.Duration
	BackupCacheExpirationTimeHours time.Duration
	AllowedEndpoints               []string
)

func init() {
	RouterAddress, _ = envkit.GetString(
		EnvProducerRouterAddress,
		DefaultProducerRouterAddress,
	)
	KibanaRouterAddress, _ = envkit.GetString(
		EnvKibanaRouterAddress,
		DefaultKibanaRouterAddress,
	)
	ProducerPort, _ = envkit.GetString(
		EnvProducerPort,
		DefaultProducerPort,
	)
	BaritoMarketUrl, _ = envkit.GetString(
		EnvBaritoMarketUrl,
		DefaultBaritoMarketUrl,
	)
	BaritoViewerUrl, _ = envkit.GetString(
		EnvBaritoViewerUrl,
		DefaultBaritoViewerUrl,
	)
	BaritoMarketAccessToken, _ = envkit.GetString(
		EnvBaritoMarketAccessToken,
		DefaultBaritoMarketAccessToken,
	)
	ProfileApiPath, _ = envkit.GetString(
		EnvBaritoProfileApiPath,
		DefaultBaritoProfileApiPath,
	)
	ProfileApiByAppGroupPath, _ = envkit.GetString(
		EnvBaritoProfileApiByAppGroupPath,
		DefaultBaritoProfileApiByAppGroupPath,
	)
	AuthorizeApiPath, _ = envkit.GetString(
		EnvBaritoAuthorizeApiPath,
		DefaultBaritoAuthorizeApiPath,
	)
	ProfileApiByClusternamePath, _ = envkit.GetString(
		EnvBaritoProfileApiByClusternamePath,
		DefaultBaritoProfileApiByClusternamePath,
	)
	JaegerServiceName, _ = envkit.GetString(
		EnvJaegerServiceName,
		DefaultJaegerServiceName,
	)
	NewRelicAppName, _ = envkit.GetString(
		EnvNewRelicAppName,
		DefaultNewRelicAppName,
	)
	NewRelicLicenseKey, _ = envkit.GetString(
		EnvNewRelicLicenseKey,
		DefaultNewRelicLicenseKey,
	)
	NewRelicEnabled, _ = envkit.GetBool(
		EnvNewRelicEnabled,
		DefaultNewRelicEnabled,
	)
	EnableTracing, _ = envkit.GetBool(
		EnvEnableTracing,
		DefaultEnableTracing,
	)
	EnableSSO, _ = envkit.GetBool(
		EnvEnableSSO,
		DefaultEnableSSO,
	)
	SSORedirectPath, _ = envkit.GetString(
		EnvSSORedirectPath,
		DefaultSSORedirectPath,
	)
	SSOClientID, _ = envkit.GetString(
		EnvSSOClientID,
		DefaultSSOClientID,
	)
	SSOClientSecret, _ = envkit.GetString(
		EnvSSOClientSecret,
		DefaultSSOClientSecret,
	)
	HMACJWTSecretString, _ = envkit.GetString(
		EnvHMACJWTSecretString,
		DefaultHMACJWTSecretString,
	)
	AllowedDomains, _ = envkit.GetString(
		EnvAllowedDomains,
		DefaultAllowedDomains,
	)
	temp, _ := envkit.GetInt(
		EnvCacheExpirationTimeSeconds,
		DefaultCacheExpirationTimeSeconds,
	)
	CacheExpirationTimeSeconds = time.Duration(temp) * time.Second

	temp, _ = envkit.GetInt(
		EnvBackupCacheExpirationTimeHours,
		DefaultBackupCacheExpirationTimeHours,
	)
	BackupCacheExpirationTimeHours = time.Duration(temp) * time.Hour

	allowedEndpointsStr, _ := envkit.GetString(
		EnvAllowedEndpoints,
		DefaultAllowedEndpoints,
	)
	AllowedEndpoints = strings.Split(allowedEndpointsStr, ",")
}
