package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	pgxPool *pgxpool.Pool
}

func (gr *GroupRepository) init(pgxPool *pgxpool.Pool) {
	gr.pgxPool = pgxPool
}

func (gr *GroupRepository) GetGroup(id int64) (entity.Group, error) {
	return selectExactlyOneFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"id": id}, nil, nil)
}

func (gr *GroupRepository) InsertGroupTx(tx pgx.Tx, args *entity.GroupArguments) (entity.Group, error) {
	return insertIntoTable[entity.Group](context.Background(), tx, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) InsertGroup(args *entity.GroupArguments) (entity.Group, error) {
	return insertIntoTable[entity.Group](context.Background(), gr.pgxPool, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) GetGroupByContentId(contentId int64) (entity.Group, error) {
	return selectExactlyOneFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"content_id": contentId}, nil, nil)
}

func (gr *GroupRepository) GetRootGroupByCanvasId(canvasId int64) (entity.Group, error) {
	return selectExactlyOneFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"canvas_id": canvasId}, nil, nil)
}
