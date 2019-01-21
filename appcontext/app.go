package appcontext

import (
	"fmt"

	"github.com/newrelic/go-agent"
)

type AppContext struct {
	newRelicApp    newrelic.Application
	newRelicConfig newrelic.Config
}

type appContextError struct {
	Error error
}

func panicIfError(err error, werr error) {
	if err != nil {
		panic(appContextError{werr})
	}
}

func NewAppContext(newRelicConfig newrelic.Config) *AppContext {
	newRelicApp, err := newrelic.NewApplication(newRelicConfig)
	panicIfError(err, fmt.Errorf("Unable to initiate NewRelic: %v", err))
	return &AppContext{
		newRelicApp:    newRelicApp,
		newRelicConfig: newRelicConfig,
	}
}

func (s *AppContext) NewRelicApp() newrelic.Application {
	return s.newRelicApp
}
