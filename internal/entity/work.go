package entity

import "time"

type Work struct {
	id        int64
	seriesId  int64
	title     string
	createdAt time.Time
}
