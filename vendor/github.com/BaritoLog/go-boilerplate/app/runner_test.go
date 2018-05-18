package app

import (
	"testing"

	"github.com/BaritoLog/go-boilerplate/errkit"
	"github.com/BaritoLog/go-boilerplate/testkit"
)

func TestRunner(t *testing.T) {
	runner := NewRunner(
		dummyContext{},
		dummyConfigManager{},
	)

	err := runner.Run()
	testkit.FatalIfError(t, err)
}

func TestRunner_config_retrive_error(t *testing.T) {
	runner := NewRunner(
		dummyContext{},
		dummyConfigManager{
			err: errkit.Error("error1"),
		},
	)

	err := runner.Run()
	testkit.FatalIfWrongError(t, err, "Config Retrive failed: error1")
}

func TestRunner_context_init_error(t *testing.T) {
	runner := NewRunner(
		dummyContext{
			initErr: errkit.Error("error1"),
		},
		dummyConfigManager{},
	)

	err := runner.Run()
	testkit.FatalIfWrongError(t, err, "Context Init failed: error1")
}

func TestRunner_context_run_error(t *testing.T) {
	runner := NewRunner(
		dummyContext{
			runErr: errkit.Error("error1"),
		},
		dummyConfigManager{},
	)

	err := runner.Run()
	testkit.FatalIfWrongError(t, err, "Context Run failed: error1")
}
