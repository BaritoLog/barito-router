package router

import "encoding/json"

type Profile struct {
	Id     string `json:"id"`
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

func (p Profile) ReceiverURL() string {
	// TODO: dig from consul
	return "https://jsonplaceholder.typicode.com/users/1"
}
