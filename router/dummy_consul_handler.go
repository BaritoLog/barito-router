package router

import "github.com/hashicorp/consul/api"

type DummyConsulHandler struct {
	catalogService *api.CatalogService
}

func (d *DummyConsulHandler) Service(consulAddr, serviceName string) (srv *api.CatalogService, err error) {
	return d.catalogService, nil
}
