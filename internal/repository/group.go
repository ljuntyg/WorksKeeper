package repository

import "github.com/jackc/pgx/v5/pgxpool"

type GroupRepository struct {
	pgxPool *pgxpool.Pool
}

func (gr *GroupRepository) Init(pgxPool *pgxpool.Pool) {
	gr.pgxPool = pgxPool
}
