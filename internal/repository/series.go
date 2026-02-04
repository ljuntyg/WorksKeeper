package repository

import (
	"WorksKeeper/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeriesRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SeriesRepository) Init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SeriesRepository) GetNSeries(n int) []*entity.Series {
	return nil
}
