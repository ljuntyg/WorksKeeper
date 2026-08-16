package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Collection struct {
	Id         int64     `db:"id"`
	InstanceId int64     `db:"instance_id"`
	CreatedAt  time.Time `db:"created_at"`
}

type CollectionArguments struct {
	InstanceId int64
}

func (ca *CollectionArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"instance_id": ca.InstanceId,
	}
}

type CollectionRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CollectionRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CollectionRepository) GetOneCollectionById(ctx context.Context, id int64) (Collection, error) {
	return selectExactlyOneFromTableWhere[Collection](ctx, cr.pgxPool, "collections",
		map[string]any{"id": id}, nil, nil)
}

func (cr *CollectionRepository) InsertCollectionTx(ctx context.Context, tx pgx.Tx, args *CollectionArguments) (Collection, error) {
	return insertIntoTable[Collection](ctx, tx, "collections", args.GetNamedArgs())
}

func (cr *CollectionRepository) GetOneCollectionByInstanceId(ctx context.Context, instanceId int64) (Collection, error) {
	return selectExactlyOneFromTableWhere[Collection](ctx, cr.pgxPool, "collections",
		map[string]any{"instance_id": instanceId}, nil, nil)
}
