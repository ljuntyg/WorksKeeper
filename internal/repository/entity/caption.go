package entity

import "github.com/jackc/pgx/v5"

type Caption struct {
	Id      int64  `db:"id"`
	MediaId int64  `db:"media_id"`
	Content string `db:"content"`
}

type CaptionArguments struct {
	MediaId int64
	Content string
}

func (ca *CaptionArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"media_id": ca.MediaId,
		"content":  ca.Content,
	}
}
