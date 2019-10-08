package router

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"

	"github.com/hashicorp/consul/api"
)

const ConsulServiceBackupCachePrefix = "consul_backup_cache_"

func consulService(consulAddr, serviceName string, cacheBag *cache.Cache) (srv *api.CatalogService, err error) {
	return fetchProducerUsingCache(cacheBag, consulAddr + "_" + serviceName, func() (srv *api.CatalogService, err error) {
		consulClient, _ := api.NewClient(&api.Config{
			Address: consulAddr,
		})

		services, _, err := consulClient.Catalog().Service(serviceName, "", nil)
		if err != nil {
			return
		}

		if len(services) < 1 {
			err = fmt.Errorf("No consul service found for '%s'", serviceName)
			return
		}

		srv = services[0]
		return
	})
}

func fetchProducerUsingCache(cacheBag *cache.Cache, key string, function func() (*api.CatalogService, error)) (srv *api.CatalogService, err error) {
	// check if still in cache
	if cacheValue, found := cacheBag.Get(key); found {
		srv = cacheValue.(*api.CatalogService)
		return
	}
	srv, err = function()

	if err == nil {
		// push to cache
		cacheBag.Set(key, srv, 1*time.Minute)
		cacheBag.Set(ConsulServiceBackupCachePrefix+key, srv, 48*time.Hour)
	} else {
		// if call is fail, check if still in backup cache
		if cacheValue, found := cacheBag.Get(ConsulServiceBackupCachePrefix + key); found {
			srv = cacheValue.(*api.CatalogService)
		}
	}

	return
}
