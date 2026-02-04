package repository

import "github.com/jackc/pgx/v5/pgxpool"

type TextRepository struct {
	pgxPool *pgxpool.Pool
}

func (tr *TextRepository) Init(pgxPool *pgxpool.Pool) {
	tr.pgxPool = pgxPool
}
