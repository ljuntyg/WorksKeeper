package entity

import "github.com/jackc/pgx/v5"

type Text struct {
	Id        int64  `db:"id"`
	ContentId int64  `db:"content_id"`
	Content   string `db:"content"`
}

type TextArguments struct {
	ContentId int64
	Content   string
}

func (ta *TextArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id": ta.ContentId,
		"content":    ta.Content,
	}
}
