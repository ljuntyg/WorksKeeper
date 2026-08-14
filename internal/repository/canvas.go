package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Canvas struct {
	Id       int64     `db:"id"`
	WorkId   int64     `db:"work_id"`
	LastEdit time.Time `db:"last_edit"`
}

// LastEdit is omitted so the column's DEFAULT NOW() applies on insert.
type CanvasArguments struct {
	WorkId int64
}

func (ca *CanvasArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"work_id": ca.WorkId,
	}
}

type CanvasRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CanvasRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CanvasRepository) GetCanvas(canvasId int64) (Canvas, error) {
	return selectExactlyOneFromTableWhere[Canvas](context.Background(), cr.pgxPool, "canvases",
		map[string]any{"id": canvasId}, nil, nil)
}

func (cr *CanvasRepository) InsertCanvas(args *CanvasArguments) (Canvas, error) {
	return insertIntoTable[Canvas](context.Background(), cr.pgxPool, "canvases", args.GetNamedArgs())
}

func (cr *CanvasRepository) InsertCanvasTx(tx pgx.Tx, args *CanvasArguments) (Canvas, error) {
	return insertIntoTable[Canvas](context.Background(), tx, "canvases", args.GetNamedArgs())
}

func (cr *CanvasRepository) GetCanvasByWorkId(workId int64) (Canvas, error) {
	return selectExactlyOneFromTableWhere[Canvas](context.Background(), cr.pgxPool, "canvases",
		map[string]any{"work_id": workId}, nil, nil)
}
