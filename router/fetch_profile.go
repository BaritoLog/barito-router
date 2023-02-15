package router

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	"github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

const ProfileBackupCachePrefix = "profile_backup_cache_"

func fetchProfileByClusterName(client *http.Client, spanContext opentracing.SpanContext, cacheBag *cache.Cache, marketUrl, accessToken, path, clusterName string) (*Profile, error) {
	return fetchUsingCache(cacheBag, accessToken+"_"+clusterName, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("access_token", accessToken)
		q.Add("cluster_name", clusterName)

		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(client, req, spanContext)
	})
}

func fetchProfileByAppSecret(client *http.Client, spanContext opentracing.SpanContext, cacheBag *cache.Cache, marketUrl, path, appSecret string) (*Profile, error) {
	return fetchUsingCache(cacheBag, appSecret, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("app_secret", appSecret)

		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(client, req, spanContext)
	})
}

func fetchProfileByAppGroupSecret(client *http.Client, spanContext opentracing.SpanContext, cacheBag *cache.Cache, marketUrl, path, appGroupSecret string, appName string) (*Profile, error) {
	return fetchUsingCache(cacheBag, appGroupSecret+"_"+appName, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("app_group_secret", appGroupSecret)
		q.Add("app_name", appName)
		req, _ := http.NewRequest("GET", address, nil)

		req.URL.RawQuery = q.Encode()

		return fetchProfile(client, req, spanContext)
	})
}

func fetchProfile(client *http.Client, req *http.Request, spanContext opentracing.SpanContext) (profile *Profile, err error) {
	if config.EnableTracing {
		err = opentracing.GlobalTracer().Inject(
			spanContext,
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			log.Errorf("Error when inject trace header: %q", err.Error())
		}
	}

	startTime := time.Now()
	res, err := client.Do(req)
	instrumentation.ObserveBaritoMarketLatency(time.Since(startTime))
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		profile, err = NewProfileFromBytes(body)
	}

	return
}

func fetchUsingCache(cacheBag *cache.Cache, key string, function func() (*Profile, error)) (profile *Profile, err error) {
	// check if still in cache
	if cacheValue, found := cacheBag.Get(key); found {
		profile = cacheValue.(*Profile)
		return
	}
	profile, err = function()

	if (err == nil) && (profile != nil) {
		// push to cache
		cacheBag.Set(key, profile, config.CacheExpirationTimeSeconds)
		cacheBag.Set(ProfileBackupCachePrefix+key, profile, config.BackupCacheExpirationTimeHours)
	} else {
		// if call is fail, check if still in backup cache
		if cacheValue, found := cacheBag.Get(ProfileBackupCachePrefix + key); found {
			profile = cacheValue.(*Profile)
		}
	}

	return
}
