package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CanvasRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CanvasRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CanvasRepository) GetCanvas(canvasId int64) (entity.Canvas, error) {
	return selectExactlyOneFromTableWhere[entity.Canvas](context.Background(), cr.pgxPool, "canvases",
		map[string]any{"id": canvasId}, nil, nil)
}

func (cr *CanvasRepository) InsertCanvas(args *entity.CanvasArguments) (entity.Canvas, error) {
	return insertIntoTable[entity.Canvas](context.Background(), cr.pgxPool, "canvases", args.GetNamedArgs())
}

func (cr *CanvasRepository) InsertCanvasTx(tx pgx.Tx, args *entity.CanvasArguments) (entity.Canvas, error) {
	return insertIntoTable[entity.Canvas](context.Background(), tx, "canvases", args.GetNamedArgs())
}

func (cr *CanvasRepository) GetCanvasByWorkId(workId int64) (entity.Canvas, error) {
	return selectExactlyOneFromTableWhere[entity.Canvas](context.Background(), cr.pgxPool, "canvases",
		map[string]any{"work_id": workId}, nil, nil)
}
