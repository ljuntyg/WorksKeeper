package entity

type Media struct {
	Id        int64  `db:"id"`
	GroupId   int64  `db:"group_id"`
	Idx       int32  `db:"idx"`
	MediaType string `db:"media_type"`
}
