package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	pgxPool *pgxpool.Pool
}

func (gr *GroupRepository) Init(pgxPool *pgxpool.Pool) {
	gr.pgxPool = pgxPool
}

func (gr *GroupRepository) InsertGroup(args *entity.GroupArguments) (entity.Group, error) {
	return insertIntoTable[entity.Group](context.Background(), gr.pgxPool, "groups", args.GetNamedArgs())
}

func (gr *GroupRepository) GetGroupsByCanvasIdWhereParentIsNullOrderByIdxAscending(canvasId int64) ([]entity.Group, error) {
	return selectFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"canvas_id": canvasId},
		&nullFilter{column: "parent_id", isNull: true},
		&orderBy{column: "idx", direction: Ascending}, nil)
}

func (gr *GroupRepository) GetGroupsByParentIdOrderByIdxAscending(parentId int64) ([]entity.Group, error) {
	return selectFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"parent_id": parentId}, nil,
		&orderBy{column: "idx", direction: Ascending}, nil)
}

func (gr *GroupRepository) GetGroupOrNilByParentIdOrderByIdxDescending(groupId int64) (*entity.Group, error) {
	return selectOptionalOneFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"parent_id": groupId}, nil, &orderBy{column: "idx", direction: Descending})
}

func (gr *GroupRepository) GetGroup(id int64) (entity.Group, error) {
	return selectExactlyOneFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"id": id}, nil, nil)
}
