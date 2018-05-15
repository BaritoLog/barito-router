package router

import "time"

type Target struct {
	consul     string
	receiver   string
	updatedAt  time.Time
	expiration time.Duration
}
