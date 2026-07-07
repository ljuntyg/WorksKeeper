package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CollectionRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CollectionRepository) GetCollection(id int64) (entity.Collection, error) {
	return selectExactlyOneFromTableWhere[entity.Collection](context.Background(), cr.pgxPool, "collections",
		map[string]any{"id": id}, nil, nil)
}

func (cr *CollectionRepository) InsertCollectionTx(tx pgx.Tx, args *entity.CollectionArguments) (entity.Collection, error) {
	return insertIntoTable[entity.Collection](context.Background(), tx, "collections", args.GetNamedArgs())
}
