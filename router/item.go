package router

import "time"

type Item struct {
	consul     string
	receiver   string
	updatedAt  time.Time
	expiration time.Duration
}
