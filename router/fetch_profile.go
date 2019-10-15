package router

import (
	"fmt"
	"github.com/BaritoLog/barito-router/config"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"net/url"
)

const ProfileBackupCachePrefix = "profile_backup_cache_"

func fetchProfileByClusterName(client *http.Client, cacheBag *cache.Cache, marketUrl, accessToken, path, clusterName string) (*Profile, error) {
	return fetchUsingCache(cacheBag, accessToken+"_"+clusterName, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("access_token", accessToken)
		q.Add("cluster_name", clusterName)

		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(client, req)
	})
}

func fetchProfileByAppSecret(client *http.Client, cacheBag *cache.Cache, marketUrl, path, appSecret string) (*Profile, error) {
	return fetchUsingCache(cacheBag, appSecret, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("app_secret", appSecret)

		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(client, req)
	})
}

func fetchProfileByAppGroupSecret(client *http.Client, cacheBag *cache.Cache, marketUrl, path, appGroupSecret string, appName string) (*Profile, error) {
	return fetchUsingCache(cacheBag, appGroupSecret+"_"+appName, func() (profile *Profile, err error) {
		address := fmt.Sprintf("%s/%s", marketUrl, path)
		q := url.Values{}
		q.Add("app_group_secret", appGroupSecret)
		q.Add("app_name", appName)
		req, _ := http.NewRequest("GET", address, nil)
		req.URL.RawQuery = q.Encode()

		return fetchProfile(client, req)
	})
}

func fetchProfile(client *http.Client, req *http.Request) (profile *Profile, err error) {
	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
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
