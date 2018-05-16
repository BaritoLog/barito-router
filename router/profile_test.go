package router

import (
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestProfile_New(t *testing.T) {
	profile := NewProfile("some-consul")
	FatalIf(t, profile.Consul() != "some-consul", "wrong consul")
}
