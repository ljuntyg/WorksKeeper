package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkRepository struct {
	pgxPool *pgxpool.Pool
}

func (wr *WorkRepository) Init(pgxPool *pgxpool.Pool) {
	wr.pgxPool = pgxPool
}

func (wr *WorkRepository) GetWork(id int64) (entity.Work, error) {
	return selectOneFromTableWhere[entity.Work](context.Background(), wr.pgxPool, "works",
		map[string]any{"id": id}, nil, nil)
}

func (wr *WorkRepository) GetNewWork(args *entity.WorkArguments) (entity.Work, error) {
	return insertIntoTable[entity.Work](context.Background(), wr.pgxPool, "works", args.GetNamedArgs())
}

func (wr *WorkRepository) GetNWorks(n int) ([]entity.Work, error) {
	return selectAllFromTableLimitN[entity.Work](context.Background(), wr.pgxPool, "works", int64(n))
}

func (wr *WorkRepository) GetWorksBySeriesId(id int64) ([]entity.Work, error) {
	return selectFromTableWhere[entity.Work](context.Background(), wr.pgxPool, "works",
		map[string]any{"series_id": id}, nil, nil)
}
