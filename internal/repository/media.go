package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
	pgxPool *pgxpool.Pool
}

func (mr *MediaRepository) Init(pgxPool *pgxpool.Pool) {
	mr.pgxPool = pgxPool
}
