package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
	pgxPool *pgxpool.Pool
}

func (mr *MediaRepository) GetMediaOrNilByGroupIdOrderByIdxDescending(groupId int64) (*entity.Media, error) {
	return selectOptionalOneFromTableWhere[entity.Media](context.Background(), mr.pgxPool, "media",
		map[string]any{"group_id": groupId}, nil, &orderBy{column: "idx", direction: Descending})
}

func (mr *MediaRepository) Init(pgxPool *pgxpool.Pool) {
	mr.pgxPool = pgxPool
}
