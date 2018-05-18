package errkit

import (
	"fmt"
	"testing"

	"github.com/BaritoLog/go-boilerplate/testkit"
)

func TestErrors(t *testing.T) {

	err1 := fmt.Errorf("err1")
	err2 := fmt.Errorf("err2")

	errors := Concat(err1, err2)

	testkit.FatalIfWrongError(t, errors, "err1: err2")

}
