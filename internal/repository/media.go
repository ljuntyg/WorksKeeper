package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	Id        int64   `db:"id"`
	ContentId int64   `db:"content_id"`
	FileHash  *string `db:"file_hash"`
	Caption   *string `db:"caption"`
}

type MediaArguments struct {
	ContentId int64
	FileHash  *string
	Caption   *string
}

func (ma *MediaArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id": ma.ContentId,
		"file_hash":  ma.FileHash,
		"caption":    ma.Caption,
	}
}

type MediaRepository struct {
	pgxPool *pgxpool.Pool
}

func (mr *MediaRepository) init(pgxPool *pgxpool.Pool) {
	mr.pgxPool = pgxPool
}

func (mr *MediaRepository) GetOneMediaById(ctx context.Context, id int64) (Media, error) {
	return selectExactlyOneFromTableWhere[Media](ctx, mr.pgxPool, "media",
		map[string]any{"id": id}, nil, nil)
}

func (mr *MediaRepository) InsertMediaTx(ctx context.Context, tx pgx.Tx, args *MediaArguments) (Media, error) {
	return insertIntoTable[Media](ctx, tx, "media", args.GetNamedArgs())
}

func (mr *MediaRepository) GetOneMediaByContentId(ctx context.Context, contentId int64) (Media, error) {
	return selectExactlyOneFromTableWhere[Media](ctx, mr.pgxPool, "media",
		map[string]any{"content_id": contentId}, nil, nil)
}

func (mr *MediaRepository) UpdateMediaSetFileHashByIdTx(ctx context.Context, tx pgx.Tx, mediaId int64, fileHash string) (Media, error) {
	return updateExactlyOneTableWhere[Media](ctx, tx, "media",
		map[string]any{"file_hash": fileHash}, map[string]any{"id": mediaId}, nil)
}

func (mr *MediaRepository) UpdateMediaSetCaptionByIdTx(ctx context.Context, tx pgx.Tx, mediaId int64, caption string) (Media, error) {
	return updateExactlyOneTableWhere[Media](ctx, tx, "media",
		map[string]any{"caption": caption}, map[string]any{"id": mediaId}, nil)
}

func (mr *MediaRepository) UpdateMediaSetCaptionNullByIdTx(ctx context.Context, tx pgx.Tx, mediaId int64) (Media, error) {
	return updateExactlyOneTableWhere[Media](ctx, tx, "media",
		map[string]any{"caption": nil}, map[string]any{"id": mediaId}, nil)
}
