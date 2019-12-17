package router

import (
	"github.com/BaritoLog/barito-router/mock"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	"github.com/patrickmn/go-cache"
	"testing"
	"time"
)

func TestConsulService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

    consulCatalog := mock.NewMockConsulCatalog(ctrl)

    returnedServices := []*api.CatalogService{
        &(api.CatalogService{ID: "test"}),
	}
	consulCatalog.EXPECT().Service(gomock.Any(), gomock.Any(), gomock.Any()).Return(returnedServices, nil, nil)

    cacheBag := cache.New(1 * time.Minute, 1 * time.Minute)
    srv, _ := fetchConsulService(consulCatalog, "someAddr", "provider", cacheBag)

	FatalIf(t, srv.ID != "test", "should return consul service")
}

func TestConsulService_noServiceAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consulCatalog := mock.NewMockConsulCatalog(ctrl)

	returnedServices := []*api.CatalogService{}
	consulCatalog.EXPECT().Service(gomock.Any(), gomock.Any(), gomock.Any()).Return(returnedServices, nil, nil)

	cacheBag := cache.New(1 * time.Minute, 1 * time.Minute)
	srv, err := fetchConsulService(consulCatalog, "someAddr", "provider", cacheBag)

	FatalIf(t, err == nil, "should return error")
	FatalIf(t, srv != nil, "should not return service")
}

func TestConsulService_withMultipleServicesAvailable(t *testing.T) {
	// if multiple service available, client should randomize the result
	// so the traffic will be spread between available service

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consulCatalog := mock.NewMockConsulCatalog(ctrl)

	returnedServices := []*api.CatalogService{
		&(api.CatalogService{ID: "test1"}),
		&(api.CatalogService{ID: "test2"}),
		&(api.CatalogService{ID: "test3"}),
	}
	consulCatalog.EXPECT().Service(gomock.Any(), gomock.Any(), gomock.Any()).Return(returnedServices, nil, nil)
	cacheBag := cache.New(1 * time.Minute, 1 * time.Minute)

	result := map[string]bool{
        "test1": false,
		"test2": false,
		"test3": false,
	}
	for i := 0;  i<30; i++ {
		srv, _ := fetchConsulService(consulCatalog, "someAddr", "provider", cacheBag)
		result[srv.ID] = true
	}

	FatalIf(t, result["test1"] == false, "service 'test1', should return at least once")
	FatalIf(t, result["test2"] == false, "service 'test2', should return at least once")
	FatalIf(t, result["test3"] == false, "service 'test3', should return at least once")
}
