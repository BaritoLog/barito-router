package router

import (
	"encoding/json"
)

type Profile struct {
	Name        string `json:"name"`
	AppGroup    string `json:"app_group"`
	TpsConfig   string `json:"tps_config"`
	ClusterName string `json:"cluster_name"`
	ConsulHost  string `json:"consul_host"`
	AppStatus   string `json:"app_status"`
}

func NewProfileFromBytes(b []byte) (*Profile, error) {
	var profile Profile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
