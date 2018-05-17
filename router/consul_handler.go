package router

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type ConsulHandler interface {
	Service(consulAddr, serviceName string) (srv *api.CatalogService, err error)
}

type consulHandler struct {
}

func NewConsulHandler() ConsulHandler {
	return &consulHandler{}
}

func (c consulHandler) Service(consulAddr, serviceName string) (srv *api.CatalogService, err error) {
	consulClient, err := api.NewClient(&api.Config{
		Address: consulAddr,
	})
	if err != nil {
		return
	}

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
}
