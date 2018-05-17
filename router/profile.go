package router

import (
	"encoding/json"
)

type Profile struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Consul string `json:"consul"`
}

func NewProfileFromBytes(b []byte) (*Profile, error) {
	var profile Profile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
