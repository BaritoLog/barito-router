package testkit

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestConsulListServiceServer(t *testing.T) {
	expect := []*api.CatalogService{
		&api.CatalogService{ID: "ID01"},
		&api.CatalogService{ID: "ID02"},
	}

	ts := NewConsulCatalogTestServer(expect)

	consulClient, _ := api.NewClient(&api.Config{
		Address: ts.URL,
	})

	got, _, err := consulClient.Catalog().Service("name", "tag", nil)
	FatalIfError(t, err)
	FatalIf(t, len(got) != len(expect), "len(got) != len(expect)")
	for i := 0; i < len(expect); i++ {
		FatalIf(t, !reflect.DeepEqual(expect[i], got[i]), "got[%d] is not equal with expect[%d]", i, i)

	}
}
