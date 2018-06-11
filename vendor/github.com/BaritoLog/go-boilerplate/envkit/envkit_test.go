package envkit

import (
	"os"
	"reflect"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestGetString(t *testing.T) {
	os.Setenv("some-key", "some-value")
	s := GetString("some-key", "default-value")

	FatalIf(t, s != "some-value", "wrong return")
}

func TestGetString_WrongKey(t *testing.T) {
	s := GetString("wrong-key", "default-value")
	FatalIf(t, s != "default-value", "wrong return")
}

func TestGetInt_WrongKey(t *testing.T) {
	i := GetInt("wrong-key", 9999)
	FatalIf(t, i != 9999, "wrong return")
}

func TestGetInt(t *testing.T) {
	os.Setenv("some-key", "8888")
	i := GetInt("some-key", 9999)

	FatalIf(t, i != 8888, "wrong return")
}

func TestGetInt_NaN(t *testing.T) {
	os.Setenv("some-key", "nan")
	i := GetInt("some-key", 9999)

	FatalIf(t, i != 9999, "wrong return")
}

func TestGetSlice_WrongKey(t *testing.T) {
	defaultSlice := []string{"1", "2"}
	slice := GetSlice("wrong-key", ",", defaultSlice)

	FatalIf(t, !reflect.DeepEqual(slice, defaultSlice), "return wrong")
}

func TestGetSlice(t *testing.T) {
	os.Setenv("some-key", "3,4,5")
	slice := GetSlice("some-key", ",", []string{"1", "2"})
	FatalIf(t, !reflect.DeepEqual(slice, []string{"3", "4", "5"}), "return wrong")

}
