package router

import (
	"fmt"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestProfile_New(t *testing.T) {
	wantClusterName := "some-cluster-name"
	wantName := "some-name"
	wantConsulHost := "some-consul-host"
	wantConsulHosts1 := "host1"
	wantConsulHosts2 := "host2"
	wantAppGroup := "some-app-group"
	wantMaxTps := 9999
	wantAppStatus := "some-app-status"

	jsonBody := fmt.Sprintf(`{
		"cluster_name": "%s",
		"name": "%s",
		"consul_host": "%s",
		"consul_hosts": ["%s", "%s"],
		"app_group_name": "%s",
		"max_tps": %d,
		"status": "%s"
	}`, wantClusterName, wantName, wantConsulHost, wantConsulHosts1, wantConsulHosts2, wantAppGroup, wantMaxTps, wantAppStatus)
	profile, err := NewProfileFromBytes([]byte(jsonBody))

	FatalIfError(t, err)
	FatalIf(t, profile.ClusterName != wantClusterName, "%s != %s", profile.ClusterName, wantClusterName)
	FatalIf(t, profile.Name != wantName, "%s != %s", profile.Name, wantName)
	FatalIf(t, profile.ConsulHost != wantConsulHost, "%s != %s", profile.ConsulHost, wantConsulHost)
	FatalIf(t, profile.ConsulHosts[0] != wantConsulHosts1, "%s != %s", profile.ConsulHosts[0], wantConsulHosts1)
	FatalIf(t, profile.ConsulHosts[1] != wantConsulHosts2, "%s != %s", profile.ConsulHosts[1], wantConsulHosts2)
	FatalIf(t, profile.AppGroup != wantAppGroup, "%s != %s", profile.AppGroup, wantAppGroup)
	FatalIf(t, profile.MaxTps != wantMaxTps, "%d != %d", profile.MaxTps, wantMaxTps)
	FatalIf(t, profile.AppStatus != wantAppStatus, "%s != %s", profile.AppStatus, wantAppStatus)
}

func TestProfile_New_InvalidJson(t *testing.T) {
	_, err := NewProfileFromBytes([]byte("invalid-json"))
	FatalIfWrongError(t, err, "invalid character 'i' looking for beginning of value")
}

func TestProfile_MetaServiceName(t *testing.T) {
	profile := Profile{
		Meta: ProfileMeta{
			ServiceNames: map[string]string{
				"service-01": "value-01",
				"service-02": "value-02",
			},
		},
	}

	val, ok := profile.MetaServiceName("wrong-service-name")
	FatalIf(t, ok, "want ok false")
	FatalIf(t, val != "", "want empty val")

	val, ok = profile.MetaServiceName("service-01")
	FatalIf(t, !ok, "want ok true")
	FatalIf(t, val != "value-01", "want val is value-01")
}

func TestProfile_MetaServiceName_NoMeta(t *testing.T) {
	profile := Profile{}
	val, ok := profile.MetaServiceName("some-service-name")
	FatalIf(t, ok, "want ok false")
	FatalIf(t, val != "", "want empty val")

}
