package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Group struct {
	Id            int64 `db:"id"`
	ContentId     int64 `db:"content_id"`
	SwapDirection bool  `db:"swap_direction"`
}

type GroupArguments struct {
	ContentId     int64
	SwapDirection bool
}

func (ga *GroupArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id":     ga.ContentId,
		"swap_direction": ga.SwapDirection,
	}
}

type GroupRepository struct {
	pgxPool *pgxpool.Pool
}

func (gr *GroupRepository) init(pgxPool *pgxpool.Pool) {
	gr.pgxPool = pgxPool
}

func (gr *GroupRepository) GetOneGroupById(ctx context.Context, id int64) (Group, error) {
	return selectExactlyOneFromTableWhere[Group](ctx, gr.pgxPool, "groups",
		map[string]any{"id": id}, nil, nil)
}

func (gr *GroupRepository) InsertGroupTx(ctx context.Context, tx pgx.Tx, args *GroupArguments) (Group, error) {
	return insertIntoTable[Group](ctx, tx, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) InsertGroup(ctx context.Context, args *GroupArguments) (Group, error) {
	return insertIntoTable[Group](ctx, gr.pgxPool, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) GetOneGroupByContentId(ctx context.Context, contentId int64) (Group, error) {
	return selectExactlyOneFromTableWhere[Group](ctx, gr.pgxPool, "groups",
		map[string]any{"content_id": contentId}, nil, nil)
}
