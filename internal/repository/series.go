package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeriesRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SeriesRepository) Init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SeriesRepository) GetSeries(id int64) (entity.Series, error) {
	return selectOneFromTableWhere[entity.Series](context.Background(), sr.pgxPool, "series",
		map[string]any{"id": id}, nil, nil)
}

func (sr *SeriesRepository) GetNewSeries(args *entity.SeriesArguments) (entity.Series, error) {
	return insertIntoTable[entity.Series](context.Background(), sr.pgxPool, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) GetNSeries(n int) ([]entity.Series, error) {
	return selectAllFromTableLimitN[entity.Series](context.Background(), sr.pgxPool, "series", int64(n))
}

func (sr *SeriesRepository) GetSeriesBySeriesId(id int64) ([]entity.Series, error) {
	return selectFromTableWhere[entity.Series](context.Background(), sr.pgxPool, "series",
		map[string]any{"parent_id": id}, nil, nil)
}
