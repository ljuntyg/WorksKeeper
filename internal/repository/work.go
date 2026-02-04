package repository

import (
	"WorksKeeper/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkRepository struct {
	pgxPool *pgxpool.Pool
}

func (wr *WorkRepository) Init(pgxPool *pgxpool.Pool) {
	wr.pgxPool = pgxPool
}

func (wr *WorkRepository) GetNWorks(n int) []*entity.Work {
	return nil
}
