package errkit

import (
	"testing"

	"github.com/BaritoLog/go-boilerplate/testkit"
)

func TestError(t *testing.T) {
	err := Error("error1")
	testkit.FatalIfWrongError(t, err, "error1")
}
