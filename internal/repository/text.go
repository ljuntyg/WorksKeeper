package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Text struct {
	Id        int64  `db:"id"`
	ContentId int64  `db:"content_id"`
	Content   string `db:"content"`
}

type TextArguments struct {
	ContentId int64
	Content   string
}

func (ta *TextArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id": ta.ContentId,
		"content":    ta.Content,
	}
}

type TextRepository struct {
	pgxPool *pgxpool.Pool
}

func (tr *TextRepository) init(pgxPool *pgxpool.Pool) {
	tr.pgxPool = pgxPool
}

func (tr *TextRepository) GetOneTextById(ctx context.Context, id int64) (Text, error) {
	return selectExactlyOneFromTableWhere[Text](ctx, tr.pgxPool, "texts",
		map[string]any{"id": id}, nil, nil)
}

func (tr *TextRepository) InsertTextTx(ctx context.Context, tx pgx.Tx, args *TextArguments) (Text, error) {
	return insertIntoTable[Text](ctx, tx, "texts", args.GetNamedArgs())
}

func (tr *TextRepository) GetOneTextByContentId(ctx context.Context, contentId int64) (Text, error) {
	return selectExactlyOneFromTableWhere[Text](ctx, tr.pgxPool, "texts",
		map[string]any{"content_id": contentId}, nil, nil)
}

func (tr *TextRepository) UpdateTextSetContentByIdTx(ctx context.Context, tx pgx.Tx, textId int64, content string) (Text, error) {
	return updateExactlyOneTableWhere[Text](ctx, tx, "texts",
		map[string]any{"content": content}, map[string]any{"id": textId}, nil)
}
