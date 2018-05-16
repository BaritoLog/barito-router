package router

type Profile interface {
	ReceiverURL() string
	Consul() string
}

type profile struct {
	consul string
}

func NewProfile(consul string) Profile {
	return &profile{consul: consul}
}

func (p profile) ReceiverURL() string {
	// TODO: dig from consul
	return "https://jsonplaceholder.typicode.com/users/1"
}

// Consul
func (p profile) Consul() string {
	return p.consul
}
