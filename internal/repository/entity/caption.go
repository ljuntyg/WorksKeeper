package entity

type Caption struct {
	Id      int64  `db:"id"`
	MediaId int64  `db:"media_id"`
	Content string `db:"content"`
}
