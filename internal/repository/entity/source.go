package entity

type Source struct {
	Id      int64  `db:"id"`
	MediaId int64  `db:"media_id"`
	Link    string `db:"link"`
}
