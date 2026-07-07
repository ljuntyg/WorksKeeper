package entity

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Series struct {
	Id        int64     `db:"id"`
	ListingId *int64    `db:"listing_id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

type SeriesArguments struct {
	ListingId *int64
	Title     string
}

func (sa *SeriesArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"listing_id": sa.ListingId,
		"title":      sa.Title,
	}
}
