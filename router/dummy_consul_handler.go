package router

import "github.com/hashicorp/consul/api"

type DummyConsulHandler struct {
	catalogService *api.CatalogService
	err            error
}

func (d *DummyConsulHandler) Service(consulAddr, serviceName string) (*api.CatalogService, error) {
	return d.catalogService, d.err
}
