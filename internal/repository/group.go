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

func (gr *GroupRepository) GetGroupsByCanvasIdWhereParentIsNullOrderByIdxAscending(canvasId int64) ([]entity.Group, error) {
	return selectFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"canvas_id": canvasId},
		&nullFilter{column: "parent_id", isNull: true},
		&orderBy{column: "idx", direction: Ascending})
}

func (gr *GroupRepository) GetGroupsByParentIdOrderByIdxAscending(parentId int64) ([]entity.Group, error) {
	return selectFromTableWhere[entity.Group](context.Background(), gr.pgxPool, "groups",
		map[string]any{"parent_id": parentId}, nil,
		&orderBy{column: "idx", direction: Ascending})
}
