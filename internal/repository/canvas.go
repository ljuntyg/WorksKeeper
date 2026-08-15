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

func (cr *CanvasRepository) GetOneCanvasById(ctx context.Context, canvasId int64) (Canvas, error) {
	return selectExactlyOneFromTableWhere[Canvas](ctx, cr.pgxPool, "canvases",
		map[string]any{"id": canvasId}, nil, nil)
}

func (cr *CanvasRepository) InsertCanvas(ctx context.Context, args *CanvasArguments) (Canvas, error) {
	return insertIntoTable[Canvas](ctx, cr.pgxPool, "canvases", args.GetNamedArgs())
}

func (cr *CanvasRepository) InsertCanvasTx(ctx context.Context, tx pgx.Tx, args *CanvasArguments) (Canvas, error) {
	return insertIntoTable[Canvas](ctx, tx, "canvases", args.GetNamedArgs())
}

func (cr *CanvasRepository) GetOneCanvasByWorkId(ctx context.Context, workId int64) (Canvas, error) {
	return selectExactlyOneFromTableWhere[Canvas](ctx, cr.pgxPool, "canvases",
		map[string]any{"work_id": workId}, nil, nil)
}
