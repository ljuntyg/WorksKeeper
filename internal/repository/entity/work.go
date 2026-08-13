package entity

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Work struct {
	Id        int64     `db:"id"`
	ListingId int64     `db:"listing_id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

type WorkArguments struct {
	ListingId int64
	Title     string
}

func (wa *WorkArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"listing_id": wa.ListingId,
		"title":      wa.Title,
	}
}
