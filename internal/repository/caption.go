package repository

import "github.com/jackc/pgx/v5/pgxpool"

type CaptionRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CaptionRepository) Init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}
