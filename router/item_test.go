package router

import (
	"testing"
	"time"

	"github.com/BaritoLog/go-boilerplate/testkit"
)

func TestItem_IsExpired_True(t *testing.T) {
	duration, _ := time.ParseDuration("-1s")
	expiredAt := time.Now().Add(duration)

	item := Item{expiredAt: expiredAt}
	testkit.FatalIf(t, item.IsExpired() == true, "Item should be expired: %+v", item)
}

func TestItem_IsExpired_False(t *testing.T) {
	duration, _ := time.ParseDuration("1s")
	expiredAt := time.Now().Add(duration)

	item := Item{expiredAt: expiredAt}
	testkit.FatalIf(t, item.IsExpired() == false, "Item should not be expired: %+v", item)
}
