package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Group struct {
	Id            int64  `db:"id"`
	CanvasId      *int64 `db:"canvas_id"`
	ContentId     *int64 `db:"content_id"`
	SwapDirection bool   `db:"swap_direction"`
}

type GroupArguments struct {
	CanvasId      *int64
	ContentId     *int64
	SwapDirection bool
}

func (ga *GroupArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"canvas_id":      ga.CanvasId,
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

func (gr *GroupRepository) GetGroup(id int64) (Group, error) {
	return selectExactlyOneFromTableWhere[Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"id": id}, nil, nil)
}

func (gr *GroupRepository) InsertGroupTx(tx pgx.Tx, args *GroupArguments) (Group, error) {
	return insertIntoTable[Group](context.Background(), tx, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) InsertGroup(args *GroupArguments) (Group, error) {
	return insertIntoTable[Group](context.Background(), gr.pgxPool, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) GetGroupByContentId(contentId int64) (Group, error) {
	return selectExactlyOneFromTableWhere[Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"content_id": contentId}, nil, nil)
}

func (gr *GroupRepository) GetRootGroupByCanvasId(canvasId int64) (Group, error) {
	return selectExactlyOneFromTableWhere[Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"canvas_id": canvasId}, nil, nil)
}
