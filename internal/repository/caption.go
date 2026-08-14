package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Caption struct {
	Id      int64  `db:"id"`
	MediaId int64  `db:"media_id"`
	Content string `db:"content"`
}

type CaptionArguments struct {
	MediaId int64
	Content string
}

func (ca *CaptionArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"media_id": ca.MediaId,
		"content":  ca.Content,
	}
}

type CaptionRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *CaptionRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *CaptionRepository) GetCaption(id int64) (Caption, error) {
	return selectExactlyOneFromTableWhere[Caption](context.Background(), cr.pgxPool, "captions",
		map[string]any{"id": id}, nil, nil)
}

func (cr *CaptionRepository) GetCaptionByMediaId(mediaId int64) (*Caption, error) {
	return selectOptionalOneFromTableWhere[Caption](context.Background(), cr.pgxPool, "captions",
		map[string]any{"media_id": mediaId}, nil, nil)
}

func (cr *CaptionRepository) InsertCaptionTx(tx pgx.Tx, args *CaptionArguments) (Caption, error) {
	return insertIntoTable[Caption](context.Background(), tx, "captions", args.GetNamedArgs())
}

func (cr *CaptionRepository) DeleteCaptionTx(tx pgx.Tx, captionId int64) error {
	return deleteFromTableWhere(context.Background(), tx, "captions",
		map[string]any{"id": captionId}, nil)
}

func (cr *CaptionRepository) UpdateCaptionContentTx(tx pgx.Tx, captionId int64, content string) (Caption, error) {
	return updateExactlyOneTableWhere[Caption](context.Background(), tx, "captions",
		map[string]any{"content": content}, map[string]any{"id": captionId}, nil)
}
