package entity

import "github.com/jackc/pgx/v5"

type Media struct {
	Id        int64  `db:"id"`
	ContentId int64  `db:"content_id"`
	CaptionId *int64 `db:"caption_id"`
	MediaType string `db:"media_type"`
}

type MediaArguments struct {
	ContentId int64
	CaptionId *int64
	MediaType string
}

func (ma *MediaArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id": ma.ContentId,
		"caption_id": ma.CaptionId,
		"media_type": ma.MediaType,
	}
}
