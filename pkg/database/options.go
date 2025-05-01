package database

import "time"

type Options func(*Postgres)

func WithConnTimeout(timeout time.Duration) Options {
	return func(c *Postgres) {
		c.conntimeout = timeout
	}
}
