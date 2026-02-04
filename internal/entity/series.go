package entity

import "time"

type Series struct {
	id        int64
	parentId  int64
	title     string
	createdAt time.Time
}
