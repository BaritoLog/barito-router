package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const ProfileBackupCachePrefix = "profile_backup_cache_"

func fetchProfileByClusterName(ctx context.Context, client *http.Client, cacheBag *cache.Cache, marketUrl, accessToken, path, clusterName string) (*Profile, error) {
	return fetchUsingCache(cacheBag, accessToken+"_"+clusterName, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("access_token", accessToken)
		q.Add("cluster_name", clusterName)

		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(ctx, client, req)
	})
}

func fetchProfileByAppSecret(ctx context.Context, client *http.Client, cacheBag *cache.Cache, marketUrl, path, appSecret string) (*Profile, error) {
	return fetchUsingCache(cacheBag, appSecret, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("app_secret", appSecret)

		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(ctx, client, req)
	})
}

func fetchProfileByAppGroupSecret(ctx context.Context, client *http.Client, cacheBag *cache.Cache, marketUrl, path, appGroupSecret string, appName string) (*Profile, error) {
	return fetchUsingCache(cacheBag, appGroupSecret+"_"+appName, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("app_group_secret", appGroupSecret)
		q.Add("app_name", appName)
		req, _ := http.NewRequest("GET", address, nil)

		req.URL.RawQuery = q.Encode()

		return fetchProfile(ctx, client, req)
	})
}

func fetchProfile(ctx context.Context, client *http.Client, req *http.Request) (profile *Profile, err error) {
	ctx, span := ProducerTracer.Start(ctx, "fetchProfile", trace.WithAttributes(
		attribute.String("url", req.URL.String()),
	))
	defer span.End()

	startTime := time.Now()
	res, err := client.Do(req)
	instrumentation.ObserveBaritoMarketLatency(time.Since(startTime))
	if err != nil {
		span.SetStatus(codes.Error, "Failed to fetch profile")
		return
	}

	if res.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		profile, err = NewProfileFromBytes(body)
		return
	}

	span.SetStatus(codes.Error, fmt.Sprintf("Failed to fetch profile, status code: %d", res.StatusCode))
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
