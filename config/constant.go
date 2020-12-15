package config

import (
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
	EnvBaritoMarketAccessToken           = "BARITO_MARKET_ACCESS_TOKEN"
	EnvBaritoProfileApiPath              = "BARITO_PROFILE_API_PATH"
	EnvBaritoProfileApiByAppGroupPath    = "BARITO_PROFILE_API_BY_APP_GROUP_PATH"
	EnvBaritoAuthorizeApiPath            = "BARITO_AUTHORIZE_API_PATH"
	EnvBaritoProfileApiByClusternamePath = "BARITO_PROFILE_API_BY_CLUSTERNAME_PATH"
	EnvCASAddress                        = "BARITO_CAS_ADDRESS"
	EnvNewRelicAppName                   = "BARITO_NEW_RELIC_APP_NAME"
	EnvNewRelicLicenseKey                = "BARITO_NEW_RELIC_LICENSE_KEY"
	EnvNewRelicEnabled                   = "BARITO_NEW_RELIC_ENABLED"
	EnvCacheExpirationTimeSeconds        = "BARITO_CACHE_EXPIRATION_TIME_IN_SECONDS"
	EnvBackupCacheExpirationTimeHours    = "BARITO_BACKUP_CACHE_EXPIRATION_TIME_IN_HOURS"
	EnvEnableTracing                     = "BARITO_ENABLE_TRACING"
	EnvJaegerServiceName                 = "BARITO_JAEGER_SERVICE_NAME"

	DefaultProducerRouterAddress             = ":8081"
	DefaultKibanaRouterAddress               = ":8083"
	DefaultProducerPort                      = ""
	DefaultBaritoMarketUrl                   = "http://localhost:3000"
	DefaultBaritoMarketAccessToken           = ""
	DefaultBaritoProfileApiPath              = "api/profile"
	DefaultBaritoProfileApiByAppGroupPath    = "api/profile_by_app_group"
	DefaultBaritoAuthorizeApiPath            = "api/authorize"
	DefaultBaritoProfileApiByClusternamePath = "api/v2/profile_by_cluster_name"
	DefaultCASAddress                        = ""
	DefaultJaegerServiceName                 = "barito_router"
	DefaultNewRelicAppName                   = "barito_router"
	DefaultNewRelicLicenseKey                = ""
	DefaultEnableTracing                     = false
	DefaultNewRelicEnabled                   = false
	DefaultCacheExpirationTimeSeconds        = 60
	DefaultBackupCacheExpirationTimeHours    = 48
)

var (
	RouterAddress                  string
	ProducerPort                   string
	KibanaRouterAddress            string
	BaritoMarketUrl                string
	BaritoMarketAccessToken        string
	ProfileApiPath                 string
	ProfileApiByAppGroupPath       string
	AuthorizeApiPath               string
	ProfileApiByClusternamePath    string
	CasAddress                     string
	JaegerServiceName              string
	NewRelicAppName                string
	NewRelicLicenseKey             string
	NewRelicEnabled                bool
	EnableTracing                  bool
	CacheExpirationTimeSeconds     time.Duration
	BackupCacheExpirationTimeHours time.Duration
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
	CasAddress, _ = envkit.GetString(
		EnvCASAddress,
		DefaultCASAddress,
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
}
