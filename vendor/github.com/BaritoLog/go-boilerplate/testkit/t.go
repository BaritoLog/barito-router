package testkit

import "fmt"

type T interface {
	Logf(format string, args ...interface{})
	FailNow()
}

type dummyT struct {
	str    string
	isFail bool
}

func (t *dummyT) Logf(format string, args ...interface{}) {
	t.str = fmt.Sprintf(format, args...)
}

func (t *dummyT) FailNow() {
	t.isFail = true
}
