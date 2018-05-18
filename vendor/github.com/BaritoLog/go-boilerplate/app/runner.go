package app

import "github.com/BaritoLog/go-boilerplate/errkit"

const (
	ErrConfigRetrieve = errkit.Error("Config Retrive failed")
	ErrContextInit    = errkit.Error("Context Init failed")
	ErrContextRun     = errkit.Error("Context Run failed")
)

// Application
type Runner interface {
	Run() (err error)
}

type runner struct {
	context       Context
	configManager ConfigurationManager
}

// NewApp create new instance of runner
func NewRunner(context Context, configManager ConfigurationManager) Runner {
	return &runner{
		context:       context,
		configManager: configManager,
	}
}

// Run the runner
func (r runner) Run() (err error) {
	var config Configuration
	config, err = r.configManager.Retrieve()
	if err != nil {
		return errkit.Concat(ErrConfigRetrieve, err)
	}

	err = r.context.Init(config)
	if err != nil {
		return errkit.Concat(ErrContextInit, err)
	}

	err = r.context.Run()
	if err != nil {
		return errkit.Concat(ErrContextRun, err)
	}

	return
}
