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

func (wr *WorkRepository) GetWork(id int64) (Work, error) {
	return selectExactlyOneFromTableWhere[Work](context.Background(), wr.pgxPool, "works",
		map[string]any{"id": id}, nil, nil)
}

func (wr *WorkRepository) InsertWork(args *WorkArguments) (Work, error) {
	return insertIntoTable[Work](context.Background(), wr.pgxPool, "works", args.GetNamedArgs())
}

func (wr *WorkRepository) InsertWorkTx(tx pgx.Tx, args *WorkArguments) (Work, error) {
	return insertIntoTable[Work](context.Background(), tx, "works", args.GetNamedArgs())
}

func (wr *WorkRepository) GetNWorks(n int) ([]Work, error) {
	return selectAllFromTableLimitN[Work](context.Background(), wr.pgxPool, "works", int64(n))
}

func (wr *WorkRepository) GetWorkByListingId(listingId int64) (Work, error) {
	return selectExactlyOneFromTableWhere[Work](context.Background(), wr.pgxPool, "works",
		map[string]any{"listing_id": listingId}, nil, nil)
}

func (wr *WorkRepository) UpdateWorkTitleTx(tx pgx.Tx, workId int64, title string) (Work, error) {
	return updateExactlyOneTableWhere[Work](context.Background(), tx, "works",
		map[string]any{"title": title}, map[string]any{"id": workId}, nil)
}
