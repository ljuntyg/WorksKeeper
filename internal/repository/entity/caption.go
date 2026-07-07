package entity

import "github.com/jackc/pgx/v5"

type Caption struct {
	Id      int64  `db:"id"`
	Content string `db:"content"`
}

type CaptionArguments struct {
	Content string
}

func (ca *CaptionArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content": ca.Content,
	}
}
