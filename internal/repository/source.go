package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SourceRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SourceRepository) init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SourceRepository) GetSourcesByMediaId(mediaId int64) ([]entity.Source, error) {
	return selectFromTableWhere[entity.Source](context.Background(), sr.pgxPool, "sources",
		map[string]any{"media_id": mediaId}, nil, nil, nil)
}
