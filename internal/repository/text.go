package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TextRepository struct {
	pgxPool *pgxpool.Pool
}

func (tr *TextRepository) init(pgxPool *pgxpool.Pool) {
	tr.pgxPool = pgxPool
}

func (tr *TextRepository) GetText(id int64) (entity.Text, error) {
	return selectExactlyOneFromTableWhere[entity.Text](context.Background(), tr.pgxPool, "texts",
		map[string]any{"id": id}, nil, nil)
}

func (tr *TextRepository) InsertTextTx(tx pgx.Tx, args *entity.TextArguments) (entity.Text, error) {
	return insertIntoTable[entity.Text](context.Background(), tx, "texts", args.GetNamedArgs())
}

func (tr *TextRepository) GetTextByContentId(contentId int64) (entity.Text, error) {
	return selectExactlyOneFromTableWhere[entity.Text](context.Background(), tr.pgxPool, "texts",
		map[string]any{"content_id": contentId}, nil, nil)
}
