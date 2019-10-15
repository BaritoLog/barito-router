package config

import (
	"github.com/BaritoLog/go-boilerplate/envkit"
	"time"
)

const (
	Name    = "Barito Router"
	Version = "0.6.2"

	EnvProducerRouterAddress             = "BARITO_PRODUCER_ROUTER"
	EnvKibanaRouterAddress               = "BARITO_KIBANA_ROUTER"
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

	DefaultProducerRouterAddress             = ":8081"
	DefaultKibanaRouterAddress               = ":8083"
	DefaultBaritoMarketUrl                   = "http://localhost:3000"
	DefaultBaritoMarketAccessToken           = ""
	DefaultBaritoProfileApiPath              = "api/profile"
	DefaultBaritoProfileApiByAppGroupPath    = "api/profile_by_app_group"
	DefaultBaritoAuthorizeApiPath            = "api/authorize"
	DefaultBaritoProfileApiByClusternamePath = "api/v2/profile_by_cluster_name"
	DefaultCASAddress                        = ""
	DefaultNewRelicAppName                   = "barito_router"
	DefaultNewRelicLicenseKey                = ""
	DefaultNewRelicEnabled                   = false
	DefaultCacheExpirationTimeSeconds        = 60
	DefaultBackupCacheExpirationTimeHours    = 48
)

var (
	RouterAddress                  string
	KibanaRouterAddress            string
	BaritoMarketUrl                string
	BaritoMarketAccessToken        string
	ProfileApiPath                 string
	ProfileApiByAppGroupPath       string
	AuthorizeApiPath               string
	ProfileApiByClusternamePath    string
	CasAddress                     string
	NewRelicAppName                string
	NewRelicLicenseKey             string
	NewRelicEnabled                bool
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
