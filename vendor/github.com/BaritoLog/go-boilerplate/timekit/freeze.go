package timekit

import (
	"time"

	"github.com/bouk/monkey"
)

var (
	patch *monkey.PatchGuard
)

func Freeze(t time.Time) {
	patch = monkey.Patch(
		time.Now,
		func() time.Time {
			return t
		},
	)
}

func FreezeUTC(s string) (err error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return
	}

	Freeze(t)
	return
}

func Unfreeze() {
	if patch != nil {
		patch.Unpatch()
	}

}
