package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Series struct {
	Id           int64     `db:"id"`
	CollectionId *int64    `db:"collection_id"`
	ListingId    *int64    `db:"listing_id"`
	Title        string    `db:"title"`
	CreatedAt    time.Time `db:"created_at"`
}

type SeriesArguments struct {
	CollectionId *int64
	ListingId    *int64
	Title        string
}

func (sa *SeriesArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"collection_id": sa.CollectionId,
		"listing_id":    sa.ListingId,
		"title":         sa.Title,
	}
}

type SeriesRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SeriesRepository) init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SeriesRepository) GetOneSeriesById(ctx context.Context, id int64) (Series, error) {
	return selectExactlyOneFromTableWhere[Series](ctx, sr.pgxPool, "series",
		map[string]any{"id": id}, nil, nil)
}

func (sr *SeriesRepository) InsertSeries(ctx context.Context, args *SeriesArguments) (Series, error) {
	return insertIntoTable[Series](ctx, sr.pgxPool, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) InsertSeriesTx(ctx context.Context, tx pgx.Tx, args *SeriesArguments) (Series, error) {
	return insertIntoTable[Series](ctx, tx, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) GetSeriesLimitN(ctx context.Context, n int) ([]Series, error) {
	return selectAllFromTableLimitN[Series](ctx, sr.pgxPool, "series", int64(n))
}

func (sr *SeriesRepository) GetOneSeriesByListingId(ctx context.Context, listingId int64) (Series, error) {
	return selectExactlyOneFromTableWhere[Series](ctx, sr.pgxPool, "series",
		map[string]any{"listing_id": listingId}, nil, nil)
}

// Only root Series carry a collection_id, so this returns the Collection's root Series.
func (sr *SeriesRepository) GetOneSeriesByCollectionId(ctx context.Context, collectionId int64) (Series, error) {
	return selectExactlyOneFromTableWhere[Series](ctx, sr.pgxPool, "series",
		map[string]any{"collection_id": collectionId}, nil, nil)
}
