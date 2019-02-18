package envkit

import (
	"os"
	"reflect"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestGetString(t *testing.T) {
	os.Setenv("some-key", "some-value")
	s, success := GetString("some-key", "default-value")

	FatalIf(t, s != "some-value" || !success, "wrong return")
}

func TestGetString_WrongKey(t *testing.T) {
	s, success := GetString("wrong-key", "default-value")
	FatalIf(t, s != "default-value" || success, "wrong return")
}

func TestGetInt_WrongKey(t *testing.T) {
	i, success := GetInt("wrong-key", 9999)
	FatalIf(t, i != 9999 || success, "wrong return")
}

func TestGetInt(t *testing.T) {
	os.Setenv("some-key", "8888")
	i, success := GetInt("some-key", 9999)

	FatalIf(t, i != 8888 || !success, "wrong return")
}

func TestGetInt_NaN(t *testing.T) {
	os.Setenv("some-key", "nan")
	i, success := GetInt("some-key", 9999)

	FatalIf(t, i != 9999 || success, "wrong return")
}

func TestGetSlice_WrongKey(t *testing.T) {
	defaultSlice := []string{"1", "2"}
	slice, success := GetSlice("wrong-key", ",", defaultSlice)

	FatalIf(t, !reflect.DeepEqual(slice, defaultSlice) || success, "return wrong")
}

func TestGetSlice(t *testing.T) {
	os.Setenv("some-key", "3,4,5")
	slice, success := GetSlice("some-key", ",", []string{"1", "2"})
	FatalIf(t, !reflect.DeepEqual(slice, []string{"3", "4", "5"}) || !success, "return wrong")

}

func TestGetBool(t *testing.T) {
	os.Setenv("some-key", "t")
	i, success := GetBool("some-key", true)

	FatalIf(t, i != true || !success, "wrong return")
}

func TestGetBool_WrongKey(t *testing.T) {
	os.Setenv("some-key", "x")
	i, success := GetBool("some-key", true)

	FatalIf(t, i != true || success, "wrong return")
}
