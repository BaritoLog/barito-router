package router

import (
	"encoding/json"
)

type Profile struct {
	Name        string       `json:"name"`
	AppGroup    string       `json:"app_group"`
	TpsConfig   string       `json:"tps_config"`
	ClusterName string       `json:"cluster_name"`
	ConsulHost  string       `json:"consul_host"`
	AppStatus   string       `json:"app_status"`
	Meta        *ProfileMeta `json:"meta"`
}

type ProfileMeta struct {
	ServiceNames map[string]string `json:"service_names"`
}

func NewProfileFromBytes(b []byte) (*Profile, error) {
	var profile Profile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (p Profile) MetaServiceName(name string) (val string, ok bool) {
	if p.Meta == nil {
		ok = false
		return
	}

	val, ok = p.Meta.ServiceNames[name]
	return
}
