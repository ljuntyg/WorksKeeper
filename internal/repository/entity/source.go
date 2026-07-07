package entity

import "github.com/jackc/pgx/v5"

type Source struct {
	Id      int64  `db:"id"`
	MediaId int64  `db:"media_id"`
	Link    string `db:"link"`
}

type SourceArguments struct {
	MediaId int64
	Link    string
}

func (sa *SourceArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"media_id": sa.MediaId,
		"link":     sa.Link,
	}
}
