package entity

type Media struct {
	Id        int64
	GroupId   int64 `db:"group_id"`
	Idx       int32
	MediaType string
}
