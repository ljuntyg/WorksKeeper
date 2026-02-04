package repository

import "github.com/jackc/pgx/v5/pgxpool"

type SourceRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SourceRepository) Init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}
