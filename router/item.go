package router

import "time"

type Item struct {
	consul    string
	receiver  string
	expiredAt time.Time
}

func (i Item) IsExpired() bool {
	return i.expiredAt.After(time.Now())
}
