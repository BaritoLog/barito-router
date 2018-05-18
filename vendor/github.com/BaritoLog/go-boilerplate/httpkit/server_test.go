package httpkit

import (
	"testing"

	"github.com/BaritoLog/go-boilerplate/errkit"
	"github.com/BaritoLog/go-boilerplate/testkit"
)

func TestDummyServer(t *testing.T) {
	server := DummyServer{
		ErrListAndServer: errkit.Error("error1"),
	}
	err := server.ListenAndServe()
	testkit.FatalIfWrongError(t, err, "error1")
}
