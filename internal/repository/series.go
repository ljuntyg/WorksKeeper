package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type SeriesRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SeriesRepository) init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SeriesRepository) GetSeries(id int64) (Series, error) {
	return selectExactlyOneFromTableWhere[Series](context.Background(), sr.pgxPool, "series",
		map[string]any{"id": id}, nil, nil)
}

func (sr *SeriesRepository) InsertSeries(args *SeriesArguments) (Series, error) {
	return insertIntoTable[Series](context.Background(), sr.pgxPool, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) InsertSeriesTx(tx pgx.Tx, args *SeriesArguments) (Series, error) {
	return insertIntoTable[Series](context.Background(), tx, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) GetNSeries(n int) ([]Series, error) {
	return selectAllFromTableLimitN[Series](context.Background(), sr.pgxPool, "series", int64(n))
}

func (sr *SeriesRepository) GetSeriesByListingId(listingId int64) (Series, error) {
	return selectExactlyOneFromTableWhere[Series](context.Background(), sr.pgxPool, "series",
		map[string]any{"listing_id": listingId}, nil, nil)
}
