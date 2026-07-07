package entity

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Collection struct {
	Id           int64     `db:"id"`
	RootSeriesId int64     `db:"root_series_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type CollectionArguments struct {
	RootSeriesId int64
}

func (ca *CollectionArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"root_series_id": ca.RootSeriesId,
	}
}
