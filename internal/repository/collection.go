package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type CollectionRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CollectionRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CollectionRepository) GetCollection(id int64) (Collection, error) {
	return selectExactlyOneFromTableWhere[Collection](context.Background(), cr.pgxPool, "collections",
		map[string]any{"id": id}, nil, nil)
}

func (cr *CollectionRepository) InsertCollectionTx(tx pgx.Tx, args *CollectionArguments) (Collection, error) {
	return insertIntoTable[Collection](context.Background(), tx, "collections", args.GetNamedArgs())
}
