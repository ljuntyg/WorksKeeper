package entity

import "time"

type Canvas struct {
	id       int64
	workId   int64
	lastEdit time.Time
}
