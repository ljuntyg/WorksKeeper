package entity

import "github.com/jackc/pgx/v5"

type Media struct {
	Id        int64  `db:"id"`
	ContentId int64  `db:"content_id"`
	MediaType string `db:"media_type"`
}

type MediaArguments struct {
	ContentId int64
	MediaType string
}

func (ma *MediaArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id": ma.ContentId,
		"media_type": ma.MediaType,
	}
}
