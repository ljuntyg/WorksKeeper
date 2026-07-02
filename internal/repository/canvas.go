package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CanvasRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CanvasRepository) Init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CanvasRepository) GetCanvasByWorkId(workId int64) (entity.Canvas, error) {
	return selectExactlyOneFromTableWhere[entity.Canvas](context.Background(), cr.pgxPool, "canvases",
		map[string]any{"work_id": workId}, nil, nil)
}

func (cr *CanvasRepository) InsertCanvas(args *entity.CanvasArguments) (entity.Canvas, error) {
	return insertIntoTable[entity.Canvas](context.Background(), cr.pgxPool, "canvases", args.GetNamedArgs())
}
