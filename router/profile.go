package router

import "time"

type Profile struct {
	consul    string
	receiver  string
	expiredAt time.Time
}

func (i Profile) IsExpired() bool {
	return i.expiredAt.After(time.Now())
}
