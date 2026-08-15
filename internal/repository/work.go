package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type WorkRepository struct {
	pgxPool *pgxpool.Pool
}

func (wr *WorkRepository) init(pgxPool *pgxpool.Pool) {
	wr.pgxPool = pgxPool
}

func (wr *WorkRepository) GetOneWorkById(ctx context.Context, id int64) (Work, error) {
	return selectExactlyOneFromTableWhere[Work](ctx, wr.pgxPool, "works",
		map[string]any{"id": id}, nil, nil)
}

func (wr *WorkRepository) InsertWork(ctx context.Context, args *WorkArguments) (Work, error) {
	return insertIntoTable[Work](ctx, wr.pgxPool, "works", args.GetNamedArgs())
}

func (wr *WorkRepository) InsertWorkTx(ctx context.Context, tx pgx.Tx, args *WorkArguments) (Work, error) {
	return insertIntoTable[Work](ctx, tx, "works", args.GetNamedArgs())
}

func (wr *WorkRepository) GetWorksLimitN(ctx context.Context, n int) ([]Work, error) {
	return selectAllFromTableLimitN[Work](ctx, wr.pgxPool, "works", int64(n))
}

func (wr *WorkRepository) GetOneWorkByListingId(ctx context.Context, listingId int64) (Work, error) {
	return selectExactlyOneFromTableWhere[Work](ctx, wr.pgxPool, "works",
		map[string]any{"listing_id": listingId}, nil, nil)
}

func (wr *WorkRepository) UpdateWorkSetTitleByIdTx(ctx context.Context, tx pgx.Tx, workId int64, title string) (Work, error) {
	return updateExactlyOneTableWhere[Work](ctx, tx, "works",
		map[string]any{"title": title}, map[string]any{"id": workId}, nil)
}
