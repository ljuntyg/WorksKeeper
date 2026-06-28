package entity

type Text struct {
	Id      int64  `db:"id"`
	GroupId int64  `db:"group_id"`
	Idx     int32  `db:"idx"`
	Content string `db:"content"`
}

func (t *Text) GetName(prefix string) string {
	return prefix + "text"
}
