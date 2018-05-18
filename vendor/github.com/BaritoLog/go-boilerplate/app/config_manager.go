package app

// Configuration
type Configuration interface {
}

// ConfigurationManager
type ConfigurationManager interface {
	Retrieve() (config Configuration, err error)
}

type dummyConfigManager struct {
	config Configuration
	err    error
}

func (m dummyConfigManager) Retrieve() (Configuration, error) {
	return m.config, m.err
}
