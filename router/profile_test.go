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
	wantAppGroup := "some-app-group"
	wantTpsConfig := "some-tps-config"
	wantAppStatus := "some-app-status"

	jsonBody := fmt.Sprintf(`{
		"cluster_name": "%s",
		"name": "%s",
		"consul_host": "%s",
		"app_group": "%s",
		"tps_config": "%s",
		"app_status": "%s"
	}`, wantClusterName, wantName, wantConsulHost, wantAppGroup, wantTpsConfig, wantAppStatus)
	profile, err := NewProfileFromBytes([]byte(jsonBody))

	FatalIfError(t, err)
	FatalIf(t, profile.ClusterName != wantClusterName, "%d != %d", profile.ClusterName, wantClusterName)
	FatalIf(t, profile.Name != wantName, "%s != %s", profile.Name, wantName)
	FatalIf(t, profile.ConsulHost != wantConsulHost, "%s != %s", profile.ConsulHost, wantConsulHost)
	FatalIf(t, profile.AppGroup != wantAppGroup, "%s != %s", profile.AppGroup, wantAppGroup)
	FatalIf(t, profile.TpsConfig != wantTpsConfig, "%s != %s", profile.TpsConfig, wantTpsConfig)
	FatalIf(t, profile.AppStatus != wantAppStatus, "%s != %s", profile.AppStatus, wantAppStatus)
}

func TestProfile_New_InvalidJson(t *testing.T) {
	_, err := NewProfileFromBytes([]byte("invalid-json"))
	FatalIfWrongError(t, err, "invalid character 'i' looking for beginning of value")
}

func TestProfile_MetaServiceName(t *testing.T) {
	profile := Profile{
		Meta: &ProfileMeta{
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
