package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TextRepository struct {
	pgxPool *pgxpool.Pool
}

func (tr *TextRepository) Init(pgxPool *pgxpool.Pool) {
	tr.pgxPool = pgxPool
}

func (tr *TextRepository) GetTextsByGroupIdOrderByIdxAscending(groupId int64) ([]entity.Text, error) {
	return selectFromTableWhere[entity.Text](context.Background(), tr.pgxPool, "texts",
		map[string]any{"group_id": groupId}, nil,
		&orderBy{column: "idx", direction: Ascending})
}
