package router

import (
	"fmt"
	"github.com/BaritoLog/barito-router/config"
	"github.com/hashicorp/consul/api"
	"github.com/patrickmn/go-cache"
	"math/rand"
	"time"
)

const ConsulServiceBackupCachePrefix = "consul_backup_cache_"

type ConsulCatalog interface {
	Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

var fetchConsulServiceFunc = fetchConsulService

func consulService(consulAddresses []string, serviceName string, cacheBag *cache.Cache) (srv *api.CatalogService, consulAddress string, err error) {
	for _, consulAddress = range consulAddresses {
		consulClient, _ := api.NewClient(&api.Config{
			Address: consulAddress,
		})

		if srv, err = fetchConsulServiceFunc(consulClient.Catalog(), consulAddress, serviceName, cacheBag); err == nil {
			return
		}
	}
	return
}

func fetchConsulService(consulCatalog ConsulCatalog, consulAddr string, serviceName string, cacheBag *cache.Cache) (srv *api.CatalogService, err error) {
	cacheKey := "services_" + consulAddr + "_" + serviceName
	services, err := fetchServicesUsingCache(cacheBag, cacheKey, func() (consulServices []*api.CatalogService, err error) {
		consulServices, _, err = consulCatalog.Service(serviceName, "", nil)
		if err != nil {
			return
		}

		if len(consulServices) < 1 {
			err = fmt.Errorf("No consul service found for '%s'", serviceName)
			return
		}

		return
	})

	if err != nil {
		return
	}

	if len(services) < 1 {
		err = fmt.Errorf("No consul service found for '%s'", serviceName)
		return
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	randomInt := r.Intn(len(services))
	srv = services[randomInt]
	return
}

func fetchServicesUsingCache(cacheBag *cache.Cache, key string, function func() ([]*api.CatalogService, error)) (services []*api.CatalogService, err error) {
	// check if still in cache
	if cacheValue, found := cacheBag.Get(key); found {
		services = cacheValue.([]*api.CatalogService)
		return
	}
	services, err = function()

	if (err == nil) && (services != nil) && (len(services) > 0) {
		// push to cache
		cacheBag.Set(key, services, config.CacheExpirationTimeSeconds)
		cacheBag.Set(ConsulServiceBackupCachePrefix+key, services, config.BackupCacheExpirationTimeHours)
	} else {
		// if call is fail, check if still in backup cache
		if cacheValue, found := cacheBag.Get(ConsulServiceBackupCachePrefix + key); found {
			services = cacheValue.([]*api.CatalogService)
		}
	}

	return
}
