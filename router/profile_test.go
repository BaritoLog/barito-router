package router

import (
	"testing"
	"time"

	"github.com/BaritoLog/go-boilerplate/testkit"
)

func TestProfile_IsExpired_True(t *testing.T) {
	duration, _ := time.ParseDuration("-1s")
	expiredAt := time.Now().Add(duration)

	item := Profile{expiredAt: expiredAt}
	testkit.FatalIf(t, item.IsExpired() == true, "Item should be expired: %+v", item)
}

func TestProfile_IsExpired_False(t *testing.T) {
	duration, _ := time.ParseDuration("1s")
	expiredAt := time.Now().Add(duration)

	item := Profile{expiredAt: expiredAt}
	testkit.FatalIf(t, item.IsExpired() == false, "Item should not be expired: %+v", item)
}
