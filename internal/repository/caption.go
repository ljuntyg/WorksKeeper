package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CaptionRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CaptionRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CaptionRepository) GetCaption(id int64) (entity.Caption, error) {
	return selectExactlyOneFromTableWhere[entity.Caption](context.Background(), cr.pgxPool, "captions",
		map[string]any{"id": id}, nil, nil)
}

func (cr *CaptionRepository) GetCaptionByMediaId(mediaId int64) (*entity.Caption, error) {
	return selectOptionalOneFromTableWhere[entity.Caption](context.Background(), cr.pgxPool, "captions",
		map[string]any{"media_id": mediaId}, nil, nil)
}
