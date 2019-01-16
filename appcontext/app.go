package appcontext

import (
	"fmt"

	"github.com/newrelic/go-agent"
)

type AppContext struct {
	newrelicApp newrelic.Application
	config      newrelic.Config
}

type appContextError struct {
	Error error
}

func panicIfError(err error, werr error) {
	if err != nil {
		panic(appContextError{werr})
	}
}

func NewAppContext(config newrelic.Config) *AppContext {
	newrelicApp, err := newrelic.NewApplication(config)
	panicIfError(err, fmt.Errorf("Unable to initiate NewRelic: %v", err))
	return &AppContext{
		newrelicApp: newrelicApp,
		config:      config,
	}
}

func (s *AppContext) NewrelicApp() newrelic.Application {
	return s.newrelicApp
}
