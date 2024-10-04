package router

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/patrickmn/go-cache"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func TestFetchProfileWithCache(t *testing.T) {

	p := Profile{
		ID:             99,
		ClusterName:    "some-cluster-name",
		Name:           "some-name",
		ConsulHosts:    []string{"some-consul-host"},
		AppGroup:       "some-app-group",
		MaxTps:         9999,
		AppStatus:      "App-status",
		AppGroupMaxTps: 18888,
		DisableAppTps:  true,
	}

	client := createClient()

	cacheBag := cache.New(1*time.Minute, 10*time.Minute)
	cacheBag.Set("app-group-secret_appname", &p, cache.DefaultExpiration)

	profile, _ := fetchProfileByAppGroupSecret(
		client,
		nil,
		cacheBag,
		"marketUrl",
		"path",
		"app-group-secret",
		"appname")

	FatalIf(t, nil == profile, "Should be able to fetch profile from cache")
	FatalIf(t, p.ClusterName != profile.ClusterName, "%s != %s", p.ClusterName, profile.ClusterName)
	FatalIf(t, p.Name != profile.Name, "%s != %s", p.Name, profile.Name)
	FatalIf(t, p.ConsulHosts[0] != profile.ConsulHosts[0], "%s != %s", p.ConsulHosts[0], profile.ConsulHosts[0])
	FatalIf(t, p.AppGroup != profile.AppGroup, "%s != %s", p.AppGroup, profile.AppGroup)
	FatalIf(t, p.MaxTps != profile.MaxTps, "%d != %d", p.MaxTps, profile.MaxTps)
	FatalIf(t, p.AppStatus != profile.AppStatus, "%s != %s", p.AppStatus, profile.AppStatus)
	FatalIf(t, p.AppGroupMaxTps != profile.AppGroupMaxTps, "%d != %d", p.AppGroupMaxTps, profile.AppGroupMaxTps)
	FatalIf(t, p.DisableAppTps != profile.DisableAppTps, "%t != %t", p.DisableAppTps, profile.DisableAppTps)
}

func TestFetchProfileWithExpiredCacheShouldCallToBaritoMarket(t *testing.T) {

	p := Profile{Name: "cached-name", AppGroup: "some-app-group"}

	cacheBag := cache.New(1*time.Minute, 10*time.Minute)
	cacheBag.Set("app-group-secret_appname", &p, 1*time.Second)
	time.Sleep(2 * time.Second)

	client := createClient()
	client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"name":"name-from-barito-market"}`)),
		}, nil
	})

	profile, _ := fetchProfileByAppGroupSecret(
		client,
		nil,
		cacheBag,
		"marketUrl",
		"path",
		"app-group-secret",
		"appname")

	FatalIf(t, "name-from-barito-market" != profile.Name, "name-from-barito-market != %s", profile.Name)
}

func TestFetchProfileWhenBaritoMarketDownShouldReturnBackupCachedProfile(t *testing.T) {

	p := Profile{Name: "long-term-cached-name", AppGroup: "some-app-group"}

	cacheBag := cache.New(1*time.Minute, 10*time.Minute)
	cacheBag.Set(ProfileBackupCachePrefix+"app-group-secret_appname", &p, 48*time.Hour)

	client := createClient()
	client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("some-error")
	})

	profile, _ := fetchProfileByAppGroupSecret(
		client,
		nil,
		cacheBag,
		"marketUrl",
		"path",
		"app-group-secret",
		"appname")

	FatalIf(t, profile == nil, "Profile should be fetched from backup cache")
	FatalIf(t, p.Name != profile.Name, "%s != %s", p.Name, profile.Name)
}
