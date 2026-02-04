package repository

import "github.com/jackc/pgx/v5/pgxpool"

type CanvasRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CanvasRepository) Init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}
