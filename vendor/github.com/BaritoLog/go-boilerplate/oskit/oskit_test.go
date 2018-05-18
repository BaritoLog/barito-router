package oskit

import (
	"os"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestGetenv(t *testing.T) {
	os.Setenv("FOO", "1")

	val := Getenv("FOO", "999")
	FatalIf(t, val != "1", "return value should be 1")

	val = Getenv("BAR", "999")
	FatalIf(t, val != "999", "return value should be 999")
}
